package eloquent

import (
	"github.com/widirahman62/pkg-go-elgo/database/query/odmquery"
)

type ODM eloquent 
type statementODM struct{*ODM}
type queryODM ODM

func (e *ODM) Where(field string) *queryODM {
	if e.query == nil {e.query = &statement{}}
	if e.query.Field != nil {
		e.query.setNilFieldStatement(&field)
		return (*queryODM)(e)
	}
	e.query.Field = &field
	return (*queryODM)(e)
}

func (e *ODM) WithOr(query1,query2 *statementODM, otherquery ...*statementODM) *statementODM{
	if e.query == nil {e.query = &statement{}}
	otherquery = append([]*statementODM{query1,query2}, otherquery...)
	query := make([]*interface{},len(otherquery))
	for i,v := range otherquery {
		query[i] = new(interface{}); *query[i] = v
	}
	if (e.query.Operator!= nil) || (e.query.Next != nil) {
		e.query.setNilOperatorValueStatement(odmquery.GetMethod(&e.Connection).Operator("or"),query)
		return &statementODM{(*ODM)(e)}
	}	
	e.query = &statement{
				Operator: odmquery.GetMethod(&e.Connection).Operator("or"),
				Value: query,
			}
	return &statementODM{(*ODM)(e)}
}

func (q *queryODM) Equals(value interface{}) *statementODM {
	if q.query.Next != nil {
		q.query.Next.setNilOperatorValueStatement(odmquery.GetMethod(&q.Connection).Operator("equal"),[]*interface{}{&value})
		return &statementODM{
			(*ODM)(q),
		}
	}
	q.query.Operator, q.query.Value = odmquery.GetMethod(&q.Connection).Operator("equal"),append([]*interface{}{}, &value)
	return &statementODM{(*ODM)(q)}
}

func (s *statementODM) setValue(val *interface{}) interface{}{
	switch v := (*val).(type) {
		case *statementODM:
			var q odmquery.Query;v.setQuery(&q)
			return q
		case *map[string]interface{}:
			imap := make(map[string]interface{}, len(*v))
			for i, item := range *v {
				imap[i] = s.setValue(&item)
			}
			return imap
		default:
			return *val
	}
}

func (s *statementODM) setQuery(o *odmquery.Query)  {
	o.Field, o.Operator = s.query.Field, s.query.Operator
	for _, v := range s.query.Value {
		newvalue :=  s.setValue(v)
		o.Value = append(o.Value, &newvalue)
	}
	if s.query.Next != nil {
		o.Next = &odmquery.Query{};(&statementODM{&ODM{query: s.query.Next}}).setQuery(o.Next)
	}
}

func (s *statementODM) Find() ([]interface{},error) {
	eloquent{Connection: s.Connection,Entity: s.Entity,Field: s.Field,FillableTag: s.FillableTag,GuardedTag: s.GuardedTag}.checkModel()
	var q odmquery.Query;s.setQuery(&q)
	return odmquery.GetMethod(&s.Connection).Find(&q,&s.Connection,&s.Entity)
}

func (e *ODM) Create(obj interface{}, otherobj ...interface{})([]interface{}, error){
	otherobj = append([]interface{}{obj}, otherobj...)
	eloquent(*e).checkObjWithModel(&e.Connection,otherobj)
	return odmquery.GetMethod(&e.Connection).Create(&e.Connection,&e.Entity,&otherobj)
}

func (s *statementODM) Update(obj interface{}, otherobj ...interface{}) (int64,error) {
	otherobj = append([]interface{}{obj}, otherobj...)
	eloquent{Connection: s.Connection,Entity: s.Entity,Field: s.Field,FillableTag: s.FillableTag,GuardedTag: s.GuardedTag}.checkObjWithModel(&s.Connection, otherobj)
	var q odmquery.Query;s.setQuery(&q)
	return odmquery.GetMethod(&s.Connection).Update(&q,&s.Connection,&s.Entity, &otherobj)
}

func (s *statementODM) Delete() (int64,error) {
	eloquent{Connection: s.Connection,Entity: s.Entity,Field: s.Field,FillableTag: s.FillableTag,GuardedTag: s.GuardedTag}.checkModel()
	var q odmquery.Query;s.setQuery(&q)
	return odmquery.GetMethod(&s.Connection).Delete(&q,&s.Connection,&s.Entity)
}
