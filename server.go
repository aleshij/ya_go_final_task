package main

// Импортируем необходимые пакеты
import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" // Драйвер для базы данных SQLite
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Agent struct {
	ID       string
	Active   bool
	LastSeen time.Time
}

// Определяем глобальные переменные
var (
	mu     sync.Mutex
	agents = make(map[string]*Agent)
	db     *sql.DB
)

// Определяем функцию init, которая инициализирует базу данных при запуске программы
func init() {
	// Открываем файл базы данных SQLite или создаем его, если он не существует
	var err error
	db, err = sql.Open("sqlite3", "./calc.db")
	if err != nil {
		log.Fatal(err)
	}
	// Создаем таблицу calc, если она не существует, с полями id, calc, stat и res
	sqlStmt := `
create table if not exists calc (id integer not null primary key, calc text, stat text, res text, agent text, t1 text, t2 text);
`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	res, err := db.Exec("UPDATE calc SET stat = 'ожидание', agent = 'нет' WHERE stat = 'в работе'")
	if err != nil {
		log.Fatal(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Изменено %d строк\n", n)

	// Создаем таблицу oper, если она не существует, с полями id, oper, symbol и time
	sqlStmt = `
create table if not exists oper (id integer not null primary key, oper text, symbol text, time integer);
`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	// Подготавливаем запрос для вставки данных в таблицу oper
	stmt, err := db.Prepare("insert into oper(id, oper, symbol, time) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	// Закрываем запрос по завершении функции
	defer stmt.Close()
	// Создаем срез с первоначальными данными для вставки
	data := [][]interface{}{
		{1, "plus", "+", 10},
		{2, "minus", "-", 10},
		{3, "multiply", "*", 10},
		{4, "divide", "/", 10},
	}
	// Подготавливаем запрос для проверки, пуста ли таблица oper
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM oper").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	// Проверяем, равно ли количество нулю
	if count == 0 {
		// Если да, значит таблица пуста, и мы можем добавить данные
		// В цикле выполняем запрос с данными из среза
		for _, row := range data {
			_, err = stmt.Exec(row...)
			if err != nil {
				log.Fatal(err)
			}
		}
		// Выводим сообщение об успешном заполнении таблицы
		log.Println("Таблица oper заполнена первоначальными данными")
	} else {
		// Если нет, значит таблица не пуста, и мы пропускаем добавление данных
		log.Println("Таблица oper не пуста, добавление данных пропущено")
	}

}

// Определяем функцию calcHandler, которая обрабатывает запросы по адресу /calc?=
func calcHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем арифметическое выражение из параметра запроса
	expr := r.URL.Query().Get("calc")
	// Проверяем, не пусто ли выражение
	if expr == "" {
		// Если да, выводим сообщение об ошибке
		fmt.Fprintln(w, "Нет выражения для вычисления")
		return
	}

	// Устанавливаем статусы
	stat := "ожидание"
	res := "нет"
	agent := "нет"
	// Время создание задачи
	// Получаем текущую дату и время
	now := time.Now()
	dat := now.Format("02.01.2006") // формат даты - день.месяц.год
	tme := now.Format("15:04:05")   // формат времени - часы:минуты:секунды
	t1 := dat + " " + tme
	t2 := "-----"

	// Подготавливаем запрос для вставки данных в таблицу calc
	stmt, err := db.Prepare("insert into calc(calc, stat, res, agent, t1, t2) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	// Закрываем запрос по завершении функции
	defer stmt.Close()
	// Выполняем запрос с переданными данными и получаем результат
	result, err := stmt.Exec(expr, stat, res, agent, t1, t2)
	if err != nil {
		log.Fatal(err)
	}
	// Получаем id добавленной строки
	id, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	// Выводим сообщение об успешном сохранении выражения и его id
	fmt.Fprintf(w, "Выражение %s успешно сохранено в базе данных.\nID выражения: %d\n", expr, id)
}

// Определяем функцию listHandler, которая обрабатывает запросы по адресу /list
func listHandler(w http.ResponseWriter, r *http.Request) {
	// Подготавливаем запрос для выборки всех данных из таблицы calc
	rows, err := db.Query("select id, calc, stat, res, agent, t1, t2 from calc")
	if err != nil {
		log.Fatal(err)
	}
	// Закрываем запрос по завершении функции
	defer rows.Close()
	// Выводим заголовок таблицы
	fmt.Fprintln(w, "| id | calc | stat | result | agent | t1 | t2 |")
	fmt.Fprintln(w, "|----|------|------|--------|-------|----|----|")
	// В цикле читаем данные из запроса
	for rows.Next() {
		// Объявляем переменные для хранения данных
		var id int
		var calc, stat, res, agent, t1, t2 string
		// Сканируем данные в переменные
		err = rows.Scan(&id, &calc, &stat, &res, &agent, &t1, &t2)
		if err != nil {
			log.Fatal(err)
		}
		// Выводим данные в виде строки таблицы
		fmt.Fprintf(w, "| %d | %s | %s | %s | %s | %s | %s |\n", id, calc, stat, res, agent, t1, t2)
	}
	// Проверяем, не произошла ли ошибка при чтении данных
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

// Определяем функцию idHandler, которая обрабатывает запросы по адресу /id=
func idHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем идентификатор выражения из параметра запроса
	id := r.URL.Query().Get("num")
	// Проверяем, не пусто ли идентификатор
	if id == "" {
		// Если да, выводим сообщение об ошибке
		fmt.Fprintln(w, "Нет идентификатора для поиска")
		return
	}
	// Подготавливаем запрос для выборки данных из таблицы calc по заданному идентификатору
	stmt, err := db.Prepare("select id, calc, stat, res, agent, t1, t2 from calc where id = ?")
	if err != nil {
		log.Fatal(err)
	}
	// Закрываем запрос по завершении функции
	defer stmt.Close()
	// Выполняем запрос с переданным идентификатором
	row := stmt.QueryRow(id)
	// Объявляем переменные для хранения данных
	var idNum int
	var calc, stat, res, agent, t1, t2 string
	// Сканируем данные в переменные
	err = row.Scan(&idNum, &calc, &stat, &res, &agent, &t1, &t2)
	// Проверяем, не произошла ли ошибка при сканировании данных
	if err != nil {
		// Если да, проверяем, не было ли это из-за отсутствия данных
		if err == sql.ErrNoRows {
			// Если да, выводим сообщение, что данные не найдены
			fmt.Fprintln(w, "Данные не найдены")
		} else {
			// Если нет, выводим сообщение об ошибке
			log.Fatal(err)
		}
		// Возвращаемся из функции
		return
	}
	// Выводим данные в виде строки
	fmt.Fprintf(w, "id: %d\ncalc: %s\nstat: %s\nres: %s\nagent: %s\nt1: %s\nt2: %s\n", idNum, calc, stat, res, agent, t1, t2)
}

// Определяем функцию operHandler, которая обрабатывает запросы по адресу /oper
func operHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем значение параметров из запроса
	plus := r.URL.Query().Get("plus")
	minus := r.URL.Query().Get("minus")
	multiply := r.URL.Query().Get("multiply")
	divide := r.URL.Query().Get("divide")
	// Проверяем, не пусто ли значение
	if plus != "" {
		// Если нет, значит запрос с параметром, и мы обновляем данные в таблице oper
		// Преобразуем значение в целое число
		time, err := strconv.Atoi(plus)
		if err != nil {
			// Если произошла ошибка, выводим сообщение об ошибке
			fmt.Fprintln(w, "Неверное значение для обновления")
			return
		}
		// Подготавливаем запрос для обновления данных в таблице oper
		stmt, err := db.Prepare("update oper set time = ? where oper = ?")
		if err != nil {
			log.Fatal(err)
		}
		// Закрываем запрос по завершении функции
		defer stmt.Close()
		// Выполняем запрос с переданными значениями
		_, err = stmt.Exec(time, "plus")
		if err != nil {
			log.Fatal(err)
		}
		// Выводим сообщение об успешном обновлении данных
		fmt.Fprintf(w, "Данные в таблице oper обновлены.\n")
	} else if minus != "" {
		// Если нет, значит запрос с параметром, и мы обновляем данные в таблице oper
		// Преобразуем значение в целое число
		time, err := strconv.Atoi(minus)
		if err != nil {
			// Если произошла ошибка, выводим сообщение об ошибке
			fmt.Fprintln(w, "Неверное значение для обновления")
			return
		}
		// Подготавливаем запрос для обновления данных в таблице oper
		stmt, err := db.Prepare("update oper set time = ? where oper = ?")
		if err != nil {
			log.Fatal(err)
		}
		// Закрываем запрос по завершении функции
		defer stmt.Close()
		// Выполняем запрос с переданными значениями
		_, err = stmt.Exec(time, "minus")
		if err != nil {
			log.Fatal(err)
		}
		// Выводим сообщение об успешном обновлении данных
		fmt.Fprintf(w, "Данные в таблице oper обновлены.\n")
	} else if multiply != "" {
		// Если нет, значит запрос с параметром, и мы обновляем данные в таблице oper
		// Преобразуем значение в целое число
		time, err := strconv.Atoi(multiply)
		if err != nil {
			// Если произошла ошибка, выводим сообщение об ошибке
			fmt.Fprintln(w, "Неверное значение для обновления")
			return
		}
		// Подготавливаем запрос для обновления данных в таблице oper
		stmt, err := db.Prepare("update oper set time = ? where oper = ?")
		if err != nil {
			log.Fatal(err)
		}
		// Закрываем запрос по завершении функции
		defer stmt.Close()
		// Выполняем запрос с переданными значениями
		_, err = stmt.Exec(time, "multiply")
		if err != nil {
			log.Fatal(err)
		}
		// Выводим сообщение об успешном обновлении данных
		fmt.Fprintf(w, "Данные в таблице oper обновлены.\n")
	} else if divide != "" {
		// Если нет, значит запрос с параметром, и мы обновляем данные в таблице oper
		// Преобразуем значение в целое число
		time, err := strconv.Atoi(divide)
		if err != nil {
			// Если произошла ошибка, выводим сообщение об ошибке
			fmt.Fprintln(w, "Неверное значение для обновления")
			return
		}
		// Подготавливаем запрос для обновления данных в таблице oper
		stmt, err := db.Prepare("update oper set time = ? where oper = ?")
		if err != nil {
			log.Fatal(err)
		}
		// Закрываем запрос по завершении функции
		defer stmt.Close()
		// Выполняем запрос с переданными значениями
		_, err = stmt.Exec(time, "divide")
		if err != nil {
			log.Fatal(err)
		}
		// Выводим сообщение об успешном обновлении данных
		fmt.Fprintf(w, "Данные в таблице oper обновлены.\n")
	} else {
		// Если да, значит запрос без параметра, и мы выводим таблицу oper
		// Подготавливаем запрос для выборки всех данных из таблицы oper
		rows, err := db.Query("select id, oper, symbol, time from oper")
		if err != nil {
			log.Fatal(err)
		}
		// Закрываем запрос по завершении функции
		defer rows.Close()
		// Выводим заголовок таблицы
		fmt.Fprintln(w, "| id | oper | symbol | time |")
		fmt.Fprintln(w, "|----|------|--------|------|")
		// В цикле читаем данные из запроса
		for rows.Next() {
			// Объявляем переменные для хранения данных
			var id, time int
			var oper, symbol string
			// Сканируем данные в переменные
			err = rows.Scan(&id, &oper, &symbol, &time)
			if err != nil {
				log.Fatal(err)
			}
			// Выводим данные в виде строки таблицы
			fmt.Fprintf(w, "| %d | %s | %s | %d |\n", id, oper, symbol, time)
		}
		// Проверяем, не произошла ли ошибка при чтении данных
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Определяем функцию main, которая запускает веб-сервер
func main() {
	// Регистрируем функцию calcHandler для обработки запросов по адресу /?calc=
	http.HandleFunc("/", calcHandler)
	// Регистрируем функцию listHandler для обработки запросов по адресу /list
	http.HandleFunc("/list", listHandler)
	// Регистрируем функцию idHandler для обработки запросов по адресу /id?num=
	http.HandleFunc("/id", idHandler)
	// Запускаем веб-сервер на порту 8080
	// Регистрируем функцию operHandler для обработки запросов по адресу /oper
	http.HandleFunc("/oper", operHandler)

	// ----
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		mu.Lock()
		agents[id] = &Agent{ID: id, Active: true, LastSeen: time.Now()}
		mu.Unlock()
	})
	http.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		for _, agent := range agents {
			fmt.Fprintf(w, "ID: %s, Active: %v, LastSeen: %s\n", agent.ID, agent.Active, agent.LastSeen)
		}
		mu.Unlock()
	})

	go checkAgents()
	// ----

	log.Println("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func checkAgents() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		mu.Lock()
		for id, agent := range agents {
			if time.Since(agent.LastSeen) > 10*time.Second {
				agent.Active = false
				_, err := db.Exec("UPDATE calc SET stat = 'ожидание', agent = 'нет' WHERE res = 'нет' AND agent = ?", id)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		mu.Unlock()
	}
}
