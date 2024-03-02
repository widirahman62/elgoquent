package odmquery

import (
	"context"
	"fmt"
	"log"

	"github.com/widirahman62/elgoquent/support/db"
)


type Query struct {
    Field    *string
    Operator *string
    Value    []*interface{}
    Next     *Query
}

type contract interface {
	Operator(string)(*string)
	Find(q *Query, conn *string, entity *string)([]interface{},error)
	Create(conn *string, entity *string, object *[]interface{}) ([]interface{}, error)
	Update(q *Query, conn *string, entity *string, object *[]interface{})(int64,error)
	Delete(q *Query, conn *string, entity *string)(int64,error)
}

type mongo struct {}

func GetMethod(conn *string) contract{
	driver := db.UseConnection(*conn).Connection["driver"]
	switch driver {
		case "mongodb":
			return &mongo{}
		default:
			log.Fatal("No supported driver in specified Connection")	
			return nil
	}
}

func (m *mongo) Operator(v string)(*string){
	var s string
	switch v {
		case "equal":
			s = "$eq"
		case "or":
			s = "$or"
	}
	return &s
}

func (m *mongo) defineBson(val *interface{}) (a interface{}){
	switch v := (*val).(type) {
	case Query:
		var bsonD db.D; m.bsonDQuery(&v, &bsonD)
		return bsonD
	case map[string]interface{}:
		bsonM := make(db.M, len(v))
		for i, item := range v {
			bsonM[i] = m.defineBson(&item)
		}
		return bsonM
	default :
		return *val
	}
}

func (m *mongo) setQueryValue(val *[]*interface{}) (a interface{}){
	if len(*val) > 1 {
		var bsonA db.A
		for _, v := range *val {
			bsonA = append(bsonA, m.defineBson(v))
		}
		return bsonA
	}
	return  m.defineBson((*val)[0])
}

func (m *mongo) bsonDQuery(q *Query, d *db.D){
	var e db.E
	if q.Field != nil {
		e.Key, e.Value = *q.Field,db.D{db.E{
				Key: *q.Operator,
				Value: m.setQueryValue(&(*q).Value),
			}}
	} else {e.Key, e.Value = *q.Operator, m.setQueryValue(&(*q).Value)}
	*d = append(*d, e)
	if q.Next != nil {
		m.bsonDQuery(q.Next,d)
	}
}

func (m *mongo) Find(q *Query, conn *string, entity *string)([]interface{},error){
	DB := db.UseConnection(*conn)
	var filter db.D; m.bsonDQuery(q,&filter)
	cursor, err := DB.MongoConnection().Database().Collection(*entity).Find(context.TODO(),DB.MongoConnection().BsonD(filter))
	if err != nil {
		return nil,fmt.Errorf(err.Error())
	}
	var data []interface{}
	if err = cursor.All(context.TODO(), &data); err != nil {
		return nil,fmt.Errorf(err.Error())
	}
	return data,nil
}

func (m *mongo) Create(conn *string, entity *string, object *[]interface{}) ([]interface{}, error) {
	DB := db.UseConnection(*conn)
	if len(*object) == 0{
		return nil,fmt.Errorf("No data to be inserted")
	}
	var data []interface{}
	if len(*object) != 1 {
		res, err := DB.MongoConnection().Database().Collection(*entity).InsertMany(context.TODO(), *object)
		for _,d := range res.InsertedIDs {
			data = append(data, d)
		}
		return data, err
	}
	res, err := DB.MongoConnection().Database().Collection(*entity).InsertOne(context.TODO(), (*object)[0]) 
	data = append(data,res.InsertedID)
	return data, err
}

func (m *mongo) Update(q *Query, conn *string, entity *string, object *[]interface{})(int64,error){
	DB := db.UseConnection(*conn)
	if len(*object) == 0{
		return 0,fmt.Errorf("No data to be updated")
	}
	var filter db.D; m.bsonDQuery(q,&filter)
	update := db.D{{"$set",DB.MongoConnection().BsonDByte(object)}}
	if len(*object) != 1 {
		res, err := DB.MongoConnection().Database().Collection(*entity).UpdateMany(context.TODO(),DB.MongoConnection().BsonD(filter),DB.MongoConnection().BsonD(update))
		return res.ModifiedCount, err
	}
	res, err := DB.MongoConnection().Database().Collection(*entity).UpdateOne(context.TODO(), DB.MongoConnection().BsonD(filter), DB.MongoConnection().BsonD(update))
	return res.ModifiedCount,err
}

func (m *mongo) Delete(q *Query, conn *string, entity *string)(int64,error){
	DB := db.UseConnection(*conn)
	var filter db.D; m.bsonDQuery(q,&filter)
	count, err := DB.MongoConnection().Database().Collection(*entity).CountDocuments(context.TODO(), DB.MongoConnection().BsonD(filter))
	switch {
		case err != nil :
			return 0,err
		case count == 0 :
			return 0,fmt.Errorf("No data to be deleted")
		case count != 1 :
			res, err := DB.MongoConnection().Database().Collection(*entity).DeleteMany(context.TODO(),DB.MongoConnection().BsonD(filter))
			return res.DeletedCount, err
		default:
			res, err := DB.MongoConnection().Database().Collection(*entity).DeleteOne(context.TODO(), DB.MongoConnection().BsonD(filter))
			return res.DeletedCount,err
	}
}
