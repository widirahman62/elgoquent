package model

import (
	"errors"
	"log"
	"reflect"
	"strings"

	"github.com/widirahman62/elgoquent/support/db"
)

var driverTag = map[string]string{
	"mongodb": "bson",
}

type statement struct {
    Field    *string
    Operator *string
    Value    []*interface{}
    Next     *statement
}

type eloquent struct {
	Connection  string
	Entity      string
	Field       interface{}
	FillableTag []string
	GuardedTag  []string
	query *statement
	*types
}

type types struct{}

func (s *statement) setNilFieldStatement(field *string) {
	if s.Next != nil {
		s.Next.setNilFieldStatement(field)
    } else {
		s.Next = &statement{
			Field: field,
		}
	}
}

func (s *statement) setNilOperatorValueStatement(op *string, val []*interface{}) {
	switch {
		case (s.Operator != nil) || (s.Next != nil):
			if s.Next == nil {
				s.Next = new(statement)
			}
			s.Next.setNilOperatorValueStatement(op, val)
		default:
			s.Operator, s.Value = op,val
	}
}

func (e eloquent) checkModel(){
	defineField(&e.Field)
	defineEntity(&e.Entity)
}

func (e eloquent) checkObjWithModel(conn *string, obj []interface{}){
	e.checkModel()
	if !checkObjWithField(&obj, &e.Field) {
		log.Fatal("Invalid object type. The object's type does not match the 'Field' member in eloquent struct.")
	}
	if driver := db.UseConnection(*conn).Connection["driver"]; driverTag[driver]==""  {
		log.Fatalf("'%s' driver not supported in elgo package", driver)	
	}

	if err := checkObjTagWithFillableGuarded(driverTag[db.UseConnection(*conn).Connection["driver"]], &obj, &e.FillableTag, &e.GuardedTag); err != nil {
		log.Fatal(err)
	}
}

func defineField(field *interface{}) {
	if field == nil {
		log.Fatal("Field in eloquent struct must be defined")
	}
}
func defineEntity(entity *string) {
	if entity == nil {
		log.Fatal("Entity in eloquent struct must be defined")
	}
}
func checkObjWithField(obj *[]interface{}, field *interface{}) bool {
	if len(*obj) == 0 {
		return false
	}
	var t = make([]reflect.Type, len(*obj))
	for i, v := range *obj {
		t[i] = reflect.TypeOf(v)
		if t[i].Kind() == reflect.Ptr {
			t[i] = t[i].Elem()
		} 
		checkObjFieldType(reflect.ValueOf(v))
	}
	for _, j := range t {
		if reflect.TypeOf(*field) != j {
			return false
		}
	}
	return true
}

func checkObjFieldType(obj reflect.Value) {
	for i := 0; i < obj.NumField(); i++ {
		if obj.Field(i).Kind() != reflect.Pointer {
			log.Fatalf("The %v member in %v must be a pointer type", obj.Type().Field(i).Name, obj.Type())
		}
	}
}

func checkObjTagWithFillableGuarded(tag string, obj *[]interface{}, fillable *[]string, guarded *[]string) error {
	var t = make([]reflect.Value, len(*obj))
	for i, v := range *obj {
		t[i] = reflect.ValueOf(v)
		if t[i].Kind() == reflect.Ptr {
			t[i] = t[i].Elem()
		}
		if t[i].IsZero() {
			return errors.New("Object is empty")
		}
		if t[i].Kind() != reflect.Struct {
			return errors.New("Invalid Object")
		}
		if err := checkTag(t[i].Type(), &tag);err != nil {
			return err
		}
	}
	fillableMap,guardedMap,err := makeMapFillableGuarded(fillable,guarded)
	if err!= nil {
        return err
    }
	for _, j := range t {
		if err = checkObjFieldItem(&tag, &j, &fillableMap, &guardedMap);err != nil {
			return err
		}
	}
	return nil
}

func makeMapFillableGuarded(fillable, guarded *[]string)(map[string]bool,map[string]bool,error){
	if fillable == nil {
		return nil,nil,errors.New("Fillable field name of struct tag cannot be nil")
	}
	if guarded == nil {
        return nil,nil,errors.New("Guarded field name of struct tag cannot be nil")
    }

	fillableMap := generateMapFromData(fillable)
	guardedMap := generateMapFromData(guarded)

	if gr,ok:=isDataInMap(guarded,fillableMap);ok{
		return nil,nil,errors.New("Fillable with '" + gr + "' field name of struct tag exist in Guarded")
	}
	if fl,ok:=isDataInMap(fillable,guardedMap);ok{
		return nil,nil,errors.New("Guarded with '" + fl + "' field name of struct tag exist in Fillable")
	}
	return fillableMap,guardedMap,nil
}

func generateMapFromData(data *[]string) map[string]bool{
	m := make(map[string]bool)
	for _, d := range *data {
		m[d] = true
	}
	return m
}

func isDataInMap(data *[]string, inMap map[string]bool)(string,bool){
	for _,d := range *data{
		if inMap[d]{
			return d,true
		}
	} 
	return "",false
}

func checkTag(obj reflect.Type, tag *string) error {
	for i := 0; i < obj.NumField(); i++ {
		if _, ok := obj.Field(i).Tag.Lookup(*tag); !ok {
			return errors.New("No " + *tag + " tag in " + obj.Field(i).Name)
		}
	}
	return nil
}

func checkObjFieldItem(tag *string, obj *reflect.Value, fillableMap *map[string]bool, guardedMap *map[string]bool) error {
	usedFields := make(map[string]bool)
	for i := 0; i < (*obj).NumField(); i++ {
		if !reflect.DeepEqual((*obj).Field(i).Interface(), reflect.Zero((*obj).Field(i).Type()).Interface()) {
			usedFields[(*obj).Type().Field(i).Name] = true
		}
	}
	for j := 0; j < (*obj).Type().NumField(); j++ {
		if usedFields[(*obj).Type().Field(j).Name] {
			return checkObjFieldTag(tag, (*obj).Type().Field(j), fillableMap, guardedMap)
		}
	}
	return nil
}

func checkObjFieldTag(tag *string, field reflect.StructField, fillableMap *map[string]bool, guardedMap *map[string]bool) error {
	fieldTag := (strings.Split(field.Tag.Get(*tag), ","))[0]
	if (*guardedMap)[fieldTag] {
		return errors.New("'" + field.Name + "' with '" + fieldTag + "' tag prevented by Guarded")
	}

	if !(*fillableMap)[fieldTag] {
		return errors.New("'" + field.Name + "' with '" + fieldTag + "' tag not registered in Fillable")
	}
	return nil
}

func (t *types) Bool(v bool) *bool {
	return &v
}

func (t *types) Byte(v byte) *byte {
	return &v
}

func (t *types) Complex128(v complex128) *complex128 {
	return &v
}

func (t *types) Complex64(v complex64) *complex64 {
	return &v
}

func (t *types) Float32(v float32) *float32 {
	return &v
}

func (t *types) Float64(v float64) *float64 {
	return &v
}

func (t *types) Int(v int) *int {
	return &v
}

func (t *types) Int8(v int8) *int8 {
	return &v
}

func (t *types) Int16(v int16) *int16 {
	return &v
}

func (t *types) Int32(v int32) *int32 {
	return &v
}

func (t *types) Int64(v int64) *int64 {
	return &v
}

func (t *types) Uint(v uint) *uint {
	return &v
}

func (t *types) Uint8(v uint8) *uint8 {
	return &v
}

func (t *types) Uint16(v uint16) *uint16 {
	return &v
}

func (t *types) Uint32(v uint32) *uint32 {
	return &v
}

func (t *types) Uint64(v uint64) *uint64 {
	return &v
}

func (t *types) String(v string) *string {
	return &v
}

func (t *types) Rune(v rune) *rune {
	return &v
}

func (t *types) Pointer(v interface{}) *interface{} {
	return &v
}
