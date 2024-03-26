package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

// Теперь нам нужно как то взаимодействовать с БД.
// Создаем переменные в которых мы будем хранить данные переменных окружения для подключению к БД.
var host = os.Getenv("HOST")
var port = os.Getenv("PORT")
var user = os.Getenv("USER")
var password = os.Getenv("PASSWORD")
var dbname = os.Getenv("DBNAME")
var sslmode = os.Getenv("SSLMODE")

var dbInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)

// Создаем таблицы в БД при подключении к ней
func createTable(tableName string, fields string) error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	if exists, err := CheckExistsTable(tableName, "public"); err == nil && !exists {
		//Создаем таблицу users
		if _, err = db.Exec(`CREATE TABLE ` + tableName + `(` + fields + `);`); err != nil {
			return err
		}
	}
	return nil
}

// Создание таблицы
func CreateTables() error {
	if err := createTable("users", "ID SERIAL PRIMARY KEY, TIMESTAMP TIMESTAMP DEFAULT CURRENT_TIMESTAMP, USERNAME TEXT, CHAT_ID INT, MESSAGE TEXT, ANSWER TEXT"); err != nil {
		return err
	}
	// Справочник криптовалют с последними ценами
	if err := createTable("dictcrypto", "ID SERIAL PRIMARY KEY, TIMESTAMP TIMESTAMP DEFAULT CURRENT_TIMESTAMP, CRYPTOID INT, CRYPTONAME TEXT, CRYPTOLASTPRICE NUMERIC(15,3), CRYPTOUPDATE TIMESTAMP"); err != nil {
		return err
	}
	// Таблица всех цен по криптовалютам
	if err := createTable("cryptoprices", "ID SERIAL PRIMARY KEY, TIMESTAMP TIMESTAMP DEFAULT CURRENT_TIMESTAMP, CRYPTOID INT, CRYPTOPRICE NUMERIC(15,3), CRYPTOUPDATE TIMESTAMP"); err != nil {
		return err
	}
	return nil
}

func CheckExistsTable(tableName string, tableSchema string) (bool, error) {
	var count uint8
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return false, err
	}
	defer db.Close()
	// Проверяем на существование таблицы в базе
	row := db.QueryRow(`select exists (select * from information_schema.tables where table_name = '` + tableName + `' and table_schema = '` + tableSchema + `')::int as "count";`)
	err = row.Scan(&count)
	if err != nil || count == 0 {
		return false, err
	}

	return true, err
}

// Таблицу мы создали, и нам нужно заносить в нее данные, этим займется следующая функция.
// Собираем данные полученные ботом
func CollectData(username string, chatid int64, message string, answer []string) error {
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Конвертируем срез с ответом в строку
	answ := strings.Join(answer, ", ")

	//Создаем SQL запрос
	data := `INSERT INTO users(username, chat_id, message, answer) VALUES($1, $2, $3, $4);`

	//Выполняем наш SQL запрос
	if _, err = db.Exec(data, `@`+username, chatid, message, answ); err != nil {
		return err
	}

	return nil
}
func WriteData(tableName string, Data map[string]string) error {
	var keys, values []string
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	//Создаем SQL запрос
	data := `INSERT INTO ` + tableName + ` (`
	for key, value := range Data {
		keys = append(keys, key)
		values = append(values, value)
	}
	keysStr := strings.Join(keys, ", ")
	valuesStr := strings.Join(values, "', '")
	data += keysStr + `) VALUES ('` + valuesStr + `');`

	//Выполняем наш SQL запрос
	if _, err = db.Exec(data); err != nil {
		return err
	}

	return nil
}

// Функция чтения из БД
// Входные данные:
// - таблица
// - отображаемые поля
// - выражения
// * нужно добавить поддержку сортировки и группировки (через интерфейсы?)
// Выходные данные: массив-интерфейс (структура), ошибка
func ReadDataRow(fields interface{}, expression []Expressions, countIter int) ([]interface{}, bool, error) {
	returnValues := []interface{}{}
	cntIter := 0
	var str string
	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return nil, false, err
	}
	defer db.Close()

	columnsPtr := getFields(fields)
	// Опредение имени колонок
	tableName, columns, err := getFieldsName(fields)
	if err != nil {
		return nil, false, err
	}
	//Создаем SQL запрос
	data := `SELECT ` + strings.Join(columns, ", ") + ` FROM ` + tableName + ` WHERE `
	for _, value := range expression {
		str += value.Join()
	}
	data += str[:len(str)-4] + `;`

	rows, err := db.Query(data)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(columnsPtr...)
		if err != nil {
			return nil, false, err
		}

		returnValue := clone(fields)

		returnValues = append(returnValues, returnValue)
		log.Println(fields)

		cntIter++
		if countIter == cntIter {
			break
		}
	}
	if err = rows.Err(); err != nil {
		return returnValues, true, err
	}
	if cntIter > 0 {
		return returnValues, true, nil
	}
	return nil, false, nil
}

// Функция определения колонок (указателей) в структуре при передаче как интерфейс
func getFields(in interface{}) (out []interface{}) {
	val := reflect.ValueOf(in).Elem()
	// Проверка на поинтер
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	// // Проверка на структуру
	// if val.Kind() != reflect.Struct {
	// 	log.Fatal("неизвестный тип")
	// }

	numCols := val.NumField()
	columns := make([]interface{}, numCols)
	for i := 0; i < numCols; i++ {
		field := val.Field(i).Addr().Interface()
		columns[i] = field
	}
	return columns
}

// Возврат имени структуры и имен полей
func getFieldsName(in interface{}) (string, []string, error) {
	val := reflect.ValueOf(in).Elem()
	structType := val.Type()
	tableName := structType.Name()
	numCols := structType.NumField()
	columns := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		field := structType.Field(i)
		fieldName := field.Name
		columns[i] = fieldName
		// println(fmt.Sprintf("%v", columns[i]))
	}

	return tableName, columns, nil
}
func clone(inter interface{}) interface{} {
	nInter := reflect.New(reflect.TypeOf(inter).Elem())

	val := reflect.ValueOf(inter).Elem()
	nVal := nInter.Elem()
	for i := 0; i < val.NumField(); i++ {
		nvField := nVal.Field(i)
		nvField.Set(val.Field(i))
	}

	return nInter.Interface()
}

// Также давайте напишем функцию которая будет считать количество уникальных пользователей
// которые писали боту, чтобы отдавать это число пользователям если они отправят боту нужную команду.
func GetNumberOfUsers() (int64, error) {
	var count int64

	//Подключаемся к БД
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Отправляем запрос в БД для подсчета числа уникальных пользователей
	row := db.QueryRow("SELECT COUNT(DISTINCT username) FROM users;")
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
