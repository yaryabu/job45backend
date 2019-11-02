package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const PersonDatabaseName = "database.txt"

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func saveToPersonDatabase(person Person) error {
	if isPersonInDatabase(person.Name) {
		return errors.New("Человек уже в базе данных!")
	}

	// Сохраняем данные в формате "имя.возвраст". То есть будет Ваня.15, Аня.10 и т.д.
	data := person.Name + "." + strconv.Itoa(person.Age) + "\n"

	err := saveToFile(PersonDatabaseName, data)

	if err != nil {
		return errors.New("Ошибка записи в базу данных!")
	}
	return nil
}

func isPersonInDatabase(name string) bool {
	person, _ := findPersonInDatabase(name)
	return person != nil
}

func findPersonInDatabase(name string) (*Person, error) {
	// Читаем файл
	bytes, err := ioutil.ReadFile(PersonDatabaseName)
	if err != nil {
		return nil, errors.New("Ошибка чтения из базы данных!")
	}
	// Делаем строку из байтов
	data := string(bytes)

	// Разделяем строку по символу перехода на новую линию и получаем массив линий
	// Пробегаемся по массиву циклом for и ищем нужного человека по имени
	for _, personString := range strings.Split(data, "\n") {
		if personString == "" {
			continue
		}
		personName := strings.Split(personString, ".")[0]
		// Нашли человека, создаем новый объект человека и возвращаем его
		if personName == name {
			personAgeString := strings.Split(personString, ".")[1]
			personAge, err := strconv.Atoi(personAgeString)
			if err != nil {
				return nil, errors.New("Не получилось конвертировать возраст человека: " + name)
			}
			return &Person{
				Name: personName,
				Age:  personAge,
			}, nil
		}
	}

	// Пробежали по всем людям, но ничего не нашли. Возвращаем ошибку
	return nil, errors.New("Человек не найден: " + name)
}

// Функция сохранения строки в файл
func saveToFile(filename, data string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := f.Write([]byte(data)); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

// Основная функция, которая отвечает за запуск программы и сервера
func main() {
	fmt.Println("Запускаем сервер...")
	http.HandleFunc("/", handleHelloWorld)
	http.HandleFunc("/createPerson", handleCreatePerson)
	http.HandleFunc("/findPerson", handleFindPerson)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

const jsonContentTypeHeader = "application/json; charset=utf-8"

// Остальные функции ниже отвечают за обработку поступления информации на сервер

func handleHelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, "Hello world!")
}

func handleCreatePerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", jsonContentTypeHeader)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var name string
	var age int
	{
		names := r.URL.Query()["name"]
		if len(names) == 0 || names[0] == "" {
			err := errors.New("Отсутствует параметр name!")
			w.WriteHeader(400)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}
		name = names[0]

		ageStrings := r.URL.Query()["age"]
		if len(ageStrings) == 0 || ageStrings[0] == "" {
			err := errors.New("Отсутствует параметр age!")
			w.WriteHeader(400)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}
		var err error
		age, err = strconv.Atoi(ageStrings[0])
		if err != nil {
			err = errors.New("Параметр age должен быть числом!")
			w.WriteHeader(400)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}
	}

	person := Person{
		Name: name,
		Age:  age,
	}
	err := saveToPersonDatabase(person)
	if err != nil {
		err = errors.New("Человек уже есть в базе данных: " + person.Name)
		w.WriteHeader(400)
		fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
		return
	}

	json.NewEncoder(w).Encode(person)
}

func handleFindPerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	name := r.URL.Query()["name"][0]
	if name == "" {
		err := errors.New("Отсутствует параметр name!")
		w.WriteHeader(400)
		fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
		return
	}

	person, err := findPersonInDatabase(name)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
		return
	}

	json.NewEncoder(w).Encode(person)
}
