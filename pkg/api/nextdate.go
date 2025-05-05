package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

// NextDate вычисляет следующую дату для задачи в соответствии с указанным правилом
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat rule cannot be empty")
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("не правильный формат даты: %v", err)
	}

	parts := strings.Split(repeat, " ")
	if len(parts) == 0 {
		return "", errors.New("invalid repeat rule format")
	}

	rule := parts[0]
	args := parts[1:]

	switch rule {
	case "d":
		if len(args) != 1 {
			return "", errors.New("invalid 'd' rule format")
		}
		interval, err := strconv.Atoi(args[0])
		if err != nil || interval < 1 || interval > 400 {
			return "", errors.New("превышено максимальное количество дней")
		}
		for {
			date = date.AddDate(0, 0, interval)
			if afterNow(date, now) {
				break
			}
		}
		return date.Format(dateFormat), nil
	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				break
			}
		}
		return date.Format(dateFormat), nil
	case "w", "m":
		return "", errors.New("значения w и m временно не поддерживается")
	default:
		return "", fmt.Errorf("unknown repeat rule: %s", rule)
	}
}

func afterNow(date, now time.Time) bool {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return date.After(now) || date.Equal(now)
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из URL
	nowStr := r.URL.Query().Get("now")
	startDate := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	var now time.Time
	var err error
	if nowStr == "" {
		now = time.Now()
	} else {
		// Парсим параметр now
		now, err = time.Parse("20060102", nowStr)
		if err != nil {
			http.Error(w, "invalid 'now' date format", http.StatusBadRequest)
			return
		}
	}
	// Вызываем вашу функцию
	result, err := NextDate(now, startDate, repeat)
	if err != nil {
		// Если ожидается пустая строка — вернем 200 с пустым телом
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}
