package cache

import (
	"WB0/internal/nats_streaming/model"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"html/template"
	"net/http"
	"time"
)

type Cache struct {
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	values            map[string]Item
}

type Item struct {
	Order      model.Order
	Created    time.Time
	Expiration int64
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := make(map[string]Item)

	cache := Cache{
		values:            items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}
	// Если интервал очистки больше 0, запускаем GC (удаление устаревших элементов)
	if cleanupInterval > 0 {
		cache.StartGC()
	}

	return &cache
}

func (c *Cache) Set(value model.Order, duration time.Duration) {

	var expiration int64
	// Если продолжительность жизни равна 0 - используется значение по-умолчанию
	if duration == 0 {
		duration = c.defaultExpiration
	}
	// Устанавливаем время истечения кеша
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	_, found := c.values[value.OrderUID]

	if found {
		return
	}

	c.values[value.OrderUID] = Item{
		Order:      value,
		Expiration: expiration,
		Created:    time.Now(),
	}

}

func (c *Cache) Get(key string) (interface{}, bool) {

	item, found := c.values[key]

	if !found {
		return nil, false
	}

	// Проверка на установку времени истечения, в противном случае он бессрочный
	if item.Expiration > 0 {

		// Если в момент запроса кеш устарел возвращаем nil
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}

	}

	return item.Order, true
}
func (c *Cache) Delete(key string) error {

	if _, found := c.values[key]; !found {
		return errors.New("Key not found")
	}

	delete(c.values, key)

	return nil
}

func (c *Cache) StartGC() {
	go c.GC()
}

func (c *Cache) GC() {

	for {
		// ожидаем время установленное в cleanupInterval
		<-time.After(c.cleanupInterval)

		if c.values == nil {
			return
		}

		// Ищем элементы с истекшим временем жизни и удаляем из хранилища
		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)

		}

	}

}

// expiredKeys возвращает список "просроченных" ключей
func (c *Cache) expiredKeys() (keys []string) {

	for k, i := range c.values {
		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
			keys = append(keys, k)
		}
	}

	return
}

// clearItems удаляет ключи из переданного списка, в нашем случае "просроченные"
func (c *Cache) clearItems(keys []string) {

	for _, k := range keys {
		delete(c.values, k)
	}
}

// PrintCache выводит все данные, находящиеся в кэше
func (c *Cache) PrintCache() {

	for _, v := range c.values {
		fmt.Println("Data in Cache: ", v.Order)
	}
}

// RestoreFromBD заполняет кэш записями из БД, если они там есть
func (c *Cache) RestoreFromBD(orders []model.Order) {

	if len(orders) == 0 {
		fmt.Println("Failed restore from BD to cache. BD is empty")
		return
	}
	for _, v := range orders {
		c.Set(v, 0)
	}
	fmt.Println("Restore from BD to cache SUCCESS")
}

func (c *Cache) HandleGetByID(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	if value, ok := c.values[id]; ok {
		responseData, err := json.Marshal(value)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)
	} else {
		http.NotFound(w, r)
	}
}

var tpl = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>

<head>
    <title>Поиск по OrderUID</title>
    <script>
        function submitForm() {
            var orderUID = document.getElementById("orderUID").value;
            window.location.href = "/order?id=" + orderUID;
            return false; // предотвращаем отправку формы на сервер
        }
    </script>
</head>

<body>
    <h1>Поиск по OrderUID</h1>
    <form onsubmit="return submitForm()">
        <label for="orderUID">Введите OrderUID:</label>
        <input type="text" id="orderUID" name="id" required>
        <button type="submit">Поиск</button>
    </form>
</body>

</html>
`))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}
