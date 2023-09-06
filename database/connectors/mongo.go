package connectors

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	config *map[string]string
}

type d interface {
	Get() *[]struct {
		Key   string
		Value interface{}
	}
}

type e interface {
	Get() *struct {
		Key   string
		Value interface{}
	}
}
type m interface {
	Get() *map[string]interface{}
}

type a interface {
	Get() *[]interface{}
}

func NewMongoDB(config *map[string]string) *MongoDB {
	conf := []string{"host", "port", "database", "username", "password"}
	for _, v := range conf {
		if _, ok := (*config)[v]; !ok {
			log.Fatalf("config %s not found", v)
		}
	}
	return &MongoDB{
		config: config,
	}
}

func (mongo *MongoDB) setBSON(data *interface{}){
	switch val := (*data).(type){
		case d:
			*data = mongo.BsonD(val)
		case e:
			*data = mongo.BsonE(val)
		case m:
			*data = mongo.BsonM(val)
		case a:
			*data = mongo.BsonA(val)
	}
}

func (m *MongoDB) BsonD(data d) bson.D {
	var bsonD bson.D
	for _, item := range *data.Get() {
		m.setBSON(&item.Value)
		bsonD = append(bsonD, bson.E{
		Key:   item.Key,
		Value: item.Value,
		})
	}
	return bsonD
}

func (m *MongoDB) BsonE(data e) bson.E {
	val := data.Get().Value
	m.setBSON(&val)
	return bson.E{
		Key:   data.Get().Key,
		Value: val,
	}
}

func (m *MongoDB) BsonM(data m) bson.M {
	for _, item := range *data.Get() {
		m.setBSON(&item)
	}
	return bson.M(*data.Get())
}

func (m *MongoDB) BsonA(data a) bson.A {
	var bsonA bson.A
	for _, item := range *data.Get() {
		m.setBSON(&item)
		bsonA = append(bsonA, item)
	}
	return bsonA
}

func (m *MongoDB) BsonDByte(a *[]interface{}) (j interface{}){
	var results []byte
	for _, input := range *a {
		value := reflect.ValueOf(input)
		if value.Kind() != reflect.Struct {
			log.Fatal("Input must be a struct")
		}
		result := make(map[string]interface{})

		for i := 0; i < value.NumField(); i++ {
			if _, ok := value.Type().Field(i).Tag.Lookup("bson"); !ok {
				log.Fatalf("No bson tag in field %v", value.Field(i))
			}
			if !reflect.DeepEqual(value.Field(i).Interface(), reflect.Zero(value.Field(i).Type()).Interface()) {
				result[strings.Split(value.Type().Field(i).Tag.Get("bson"), ",")[0]] = value.Field(i).Interface()
			}
		}

		jsonResult, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("can't marshal: %s\n", err)
		}

		results = append(results, jsonResult...)
	}
	
	err := bson.UnmarshalExtJSON(results,true,&j)
	if err !=nil {
		log.Fatalf("can't marshal: %s\n", err)
	}
	return j
}


func (m *MongoDB) client() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI((m).getDsn(m.config)))
	if err != nil {
		log.Fatal(err)
	}
	(m).ping(ctx, client)
	return client
}

func (m *MongoDB) Database() *mongo.Database {
	return (m).client().Database((*m.config)["database"])
}

func (m *MongoDB) Version() string {
	var commandResult bson.M
	command := bson.D{{Key: "buildInfo", Value: 1}}
	err := (m).Database().RunCommand(context.Background(), command).Decode(&commandResult)
	if err != nil {
		log.Fatal(err)
	}
	return commandResult["version"].(string)
}

func (m *MongoDB) getDsn(config *map[string]string) string {
	if (*config)["unix-socket"] != "" {
		return (m).getSocketDsn(*&config)
	}
	return (m).getHostDsn(*&config)

}

func (m *MongoDB) ping(ctx context.Context, client *mongo.Client) {
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
}

func (m *MongoDB) getSocketDsn(config *map[string]string) string {
	return "mongodb+srv://" + (*config)["username"] + ":" + (*config)["password"] + "@unix" + (*config)["socket"] + ")/" + (*config)["database"] + "?retryWrites=true&w=majority"
}

func (m *MongoDB) getHostDsn(config *map[string]string) string {
	return "mongodb://" + (*config)["username"] + ":" + (*config)["password"] + "@" + (*config)["host"] + ":" + (*config)["port"] + "/" + (*config)["database"]
}
