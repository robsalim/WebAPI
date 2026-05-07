package services

import (
    "database/sql"
    "fmt"
    "regexp"
    "strings"
    "time"
    "unicode"
	"context"
    "webapi/config"
    "webapi/models"
	//"log"
	_ "github.com/denisenkom/go-mssqldb"
)

type DatabaseHealthService struct {
	cfg *config.Config
}

func NewDatabaseHealthService(cfg *config.Config) *DatabaseHealthService {
	return &DatabaseHealthService{cfg: cfg}
}

func (s *DatabaseHealthService) CheckDatabaseHealthAsync() (*models.DatabaseHealthResponse, error) {
	// Добавляем таймаут в строку подключения
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;timeout=5;dial timeout=5",
		s.cfg.Database.DataSource,
		s.cfg.Database.User,
		s.cfg.Database.Password,
		s.cfg.Database.Base,
	)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return &models.DatabaseHealthResponse{
			Success:       false,
			SqlConnection: fmt.Sprintf("❌ Error: %v", err),
			XmlStatus:     "N/A",
			LastCheck:     time.Now(),
		}, nil
	}
	defer db.Close()

	// Контекст с таймаутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return &models.DatabaseHealthResponse{
			Success:       false,
			SqlConnection: fmt.Sprintf("❌ Error: %v", err),
			XmlStatus:     "N/A",
			LastCheck:     time.Now(),
		}, nil
	}

	xmlStatus := s.getXmlStatus(db)

	return &models.DatabaseHealthResponse{
		Success:       true,
		SqlConnection: "✅ Connected",
		XmlStatus:     xmlStatus,
		LastCheck:     time.Now(),
	}, nil
}

func (s *DatabaseHealthService) CheckDataDelaysAsync() (*models.DataDelaysResponse, error) {
	// Добавляем таймаут в строку подключения
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;timeout=5;dial timeout=5",
		s.cfg.Database.DataSource,
		s.cfg.Database.User,
		s.cfg.Database.Password,
		s.cfg.Database.Base,
	)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return &models.DataDelaysResponse{
			Success:            false,
			HasDelays:          false,
			DelayedPointsCount: 0,
			DelayedPoints:      []models.DelayedPoint{},
			Error:              err.Error(),
			LastCheck:          time.Now(),
		}, nil
	}
	defer db.Close()

	// Контекст с таймаутом 10 секунд (для сложного запроса)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Проверяем соединение с таймаутом
	err = db.PingContext(ctx)
	if err != nil {
		return &models.DataDelaysResponse{
			Success:            false,
			HasDelays:          false,
			DelayedPointsCount: 0,
			DelayedPoints:      []models.DelayedPoint{},
			Error:              fmt.Sprintf("DB connection error: %v", err),
			LastCheck:          time.Now(),
		}, nil
	}

	delayedPoints := s.getDelayedPointsWithTimeout(ctx, db)
	hasDelays := len(delayedPoints) > 0

	return &models.DataDelaysResponse{
		Success:            true,
		HasDelays:          hasDelays,
		DelayedPointsCount: len(delayedPoints),
		DelayedPoints:      delayedPoints,
		LastCheck:          time.Now(),
	}, nil
}

func (s *DatabaseHealthService) getXmlStatus(db *sql.DB) string {
	// Контекст с таймаутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := "SELECT TOP 1 [MESSAGE] FROM XMLMESSAGE ORDER BY [TIME] DESC"
	var xmlMessage string
	
	err := db.QueryRowContext(ctx, query).Scan(&xmlMessage)
	if err != nil {
		if err == sql.ErrNoRows {
			return "❌ Not found"
		}
		return fmt.Sprintf("❌ Error: %v", err)
	}

	return s.parseXmlDateStatus(xmlMessage)
}
/*
func (s *DatabaseHealthService) getDelayedPointsWithTimeout(ctx context.Context, db *sql.DB) []models.DelayedPoint {
	var delayedPoints []models.DelayedPoint

	query := "SELECT [LASTDATE], [LASTHOUR], [NAME] FROM CONTROL WHERE STOPED = 'false'"
	
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return delayedPoints
	}
	defer rows.Close()

	for rows.Next() {
		var lastDate time.Time
		var lastHour int
		var name string

		err := rows.Scan(&lastDate, &lastHour, &name)
		if err != nil {
			continue
		}

		fullDate := lastDate.Add(time.Duration(lastHour) * time.Hour)
		delayHours := time.Since(fullDate).Hours()
		
		// В функции getDelayedPointsWithTimeout и getDelayedPoints
		if delayHours > 0.0{; //s.cfg.Delays.DeltaHour {  // теперь float64 сравнивается с float64
			delayedPoints = append(delayedPoints, models.DelayedPoint{
				Name:       name,
				LastDate:   fullDate,
				DelayHours: roundFloat(delayHours, 2),
			})
		}
		
		
	}

	return delayedPoints
}
*/

func (s *DatabaseHealthService) getDelayedPointsWithTimeout(ctx context.Context, db *sql.DB) []models.DelayedPoint {
	var delayedPoints []models.DelayedPoint

	query := "SELECT [LASTDATE], [LASTHOUR], [NAME] FROM CONTROL WHERE STOPED = 'false'"
	
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return delayedPoints
	}
	defer rows.Close()

	for rows.Next() {
		var lastDate time.Time
		var lastHour int
		var name string

		rows.Scan(&lastDate, &lastHour, &name)

		// Создаем время: дата из БД + час из БД
		fullDate := time.Date(lastDate.Year(), lastDate.Month(), lastDate.Day(), lastHour, 0, 0, 0, time.Local)
		delayHours := time.Since(fullDate).Hours()

		if delayHours > s.cfg.Delays.DeltaHour {
			delayedPoints = append(delayedPoints, models.DelayedPoint{
				Name:       name,
				LastDate:   fullDate,
				DelayHours: roundFloat(delayHours, 2),
			})
		}
	}

	return delayedPoints
}
func (s *DatabaseHealthService) parseXmlDateStatus(xmlMessage string) string {
	// 1. Удаляем все непечатные символы
	clean := strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\r' || r == '\t' {
			return r
		}
		return -1
	}, xmlMessage)

	// 2. Ищем русскую дату вида "3 мая 2026 г." или "03 мая 2026 г."
	re := regexp.MustCompile(`(\d{1,2})\s+([а-я]+)\s+(\d{4})\s+г\.`)
	matches := re.FindStringSubmatch(clean)
	if len(matches) >= 4 {
		day := matches[1]
		monthRu := strings.ToLower(matches[2])
		year := matches[3]

		// Маппинг русских месяцев на номера
		months := map[string]string{
			"января": "01", "февраля": "02", "марта": "03", "апреля": "04",
			"мая": "05", "июня": "06", "июля": "07", "августа": "08",
			"сентября": "09", "октября": "10", "ноября": "11", "декабря": "12",
		}
		monthNum, ok := months[monthRu]
		if !ok {
			return fmt.Sprintf("❌ Неизвестный месяц: %s", monthRu)
		}

		// Формируем дату с ведущими нулями
		dayPadded := fmt.Sprintf("%02s", day)
		dateStr := fmt.Sprintf("%s.%s.%s", dayPadded, monthNum, year)
		
		parsedDate, err := time.Parse("02.01.2006", dateStr)
		if err != nil {
			return fmt.Sprintf("❌ Ошибка парсинга: %v", err)
		}

		timeDiff := time.Since(parsedDate)
		daysDiff := int(timeDiff.Hours() / 24)
		if daysDiff < 2 {
			return fmt.Sprintf("✅ Отправлено за (%s)", parsedDate.Format("02.01.2006"))
		}
		return fmt.Sprintf("❌ Ошибка, последняя отправка (%s) — %d дней назад", parsedDate.Format("02.01.2006"), daysDiff)
	}

	// 3. Пробуем стандартные форматы (с поддержкой дня без ведущего нуля)
	formats := []string{
		"2.01.2006",
		"02.01.2006",
		"2.1.2006",
		"02.1.2006",
		"2.01.2006 15:04:05",
		"02.01.2006 15:04:05",
		"2006-01-02",
		"2006-1-2",
	}
	for _, format := range formats {
		parsedDate, err := time.Parse(format, strings.TrimSpace(clean))
		if err == nil {
			timeDiff := time.Since(parsedDate)
			daysDiff := int(timeDiff.Hours() / 24)
			if daysDiff < 2 {
				return fmt.Sprintf("✅ Отправлено за (%s)", parsedDate.Format("02.01.2006 15:04"))
			}
			return fmt.Sprintf("❌ Ошибка, последняя отправка (%s) — %d дней назад", parsedDate.Format("02.01.2006 15:04"), daysDiff)
		}
	}

	// 4. Если не распарсили, показываем отладочную информацию
	preview := clean
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	return fmt.Sprintf("❌ Не удалось распознать дату: %s", preview)
}

func (s *DatabaseHealthService) getDelayedPoints(db *sql.DB) []models.DelayedPoint {
	var delayedPoints []models.DelayedPoint

	query := "SELECT [LASTDATE], [LASTHOUR], [NAME] FROM CONTROL WHERE STOPED = 'false'"
	rows, err := db.Query(query)
	if err != nil {
		return delayedPoints
	}
	defer rows.Close()

	for rows.Next() {
		var lastDate time.Time
		var lastHour int
		var name string

		err := rows.Scan(&lastDate, &lastHour, &name)
		if err != nil {
			continue
		}

		fullDate := lastDate.Add(time.Duration(lastHour) * time.Hour)
		delayHours := time.Since(fullDate).Hours()

		// В функции getDelayedPointsWithTimeout и getDelayedPoints
		if delayHours > s.cfg.Delays.DeltaHour {  // теперь float64 сравнивается с float64
			delayedPoints = append(delayedPoints, models.DelayedPoint{
				Name:       name,
				LastDate:   fullDate,
				DelayHours: roundFloat(delayHours, 2),
			})
		}
		
	}

	return delayedPoints
}

func roundFloat(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio)) / ratio
}