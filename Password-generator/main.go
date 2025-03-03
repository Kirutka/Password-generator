package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type PageData struct {
    Password string
}

func generatePassword(length int, uppercase, numbers, symbols bool) string {
    const (
        lowercaseChars = "abcdefghijklmnopqrstuvwxyz"
        uppercaseChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
        numberChars    = "0123456789"
        symbolChars    = "!@#$%^&*()-_=+[]{}|;:,.<>?"
    )

    // Определяем доступные символы
    var chars string
    if uppercase {
        chars += uppercaseChars
    }
    if numbers {
        chars += numberChars
    }
    if symbols {
        chars += symbolChars
    }
    chars += lowercaseChars // Всегда включаем строчные буквы

    // Генерация пароля
    rand.Seed(time.Now().UnixNano())
    password := make([]byte, length)
    for i := 0; i < length; i++ {
        password[i] = chars[rand.Intn(len(chars))]
    }

    return string(password)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("index.html"))
    tmpl.Execute(w, nil)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        r.ParseForm()

        // Получаем параметры из формы
        lengthStr := r.Form.Get("length")
        length, err := strconv.Atoi(lengthStr)
        if err != nil || length < 4 || length > 32 {
            http.Error(w, "Неверная длина пароля", http.StatusBadRequest)
            return
        }

        uppercase := r.Form.Get("uppercase") == "on"
        numbers := r.Form.Get("numbers") == "on"
        symbols := r.Form.Get("symbols") == "on"

        // Генерируем пароль
        password := generatePassword(length, uppercase, numbers, symbols)

        // Отображаем результат
        data := PageData{Password: password}
        tmpl := template.Must(template.ParseFiles("index.html"))
        tmpl.Execute(w, data)
    }
}

func main() {
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/generate", generateHandler)

    fmt.Println("Сервер запущен на http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}