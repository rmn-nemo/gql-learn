/*
Сервер создан в тестовых целях для изчуния GraphQL,
в частности запросов на объекты, имеющих поле, возвращающее значения разных типов.

Пример запроса:
{
	events {
		id
		name
		payload {
			... on Document {
				id
				title
			}
			... on Report {
				id
				name
			}
		}
	}
}
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/graphql-go/graphql"
	"net/http"
)

const (
	ServerHost string = "localhost"
	ServerPort string = "8080"
)

const (
	ReportCreatedEvent string = "REPORT_CREATED_EVENT"
	FileUploadedEvent  string = "FILE_UPLOADED_EVENT"
)

type RequestBody struct {
	Query string `json:"query"`
}

type Event struct {
	ID      string      `json:"id,omitempty"`
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Document struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title"`
}

type Report struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

var documentType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Document",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var reportType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Report",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var unionType = graphql.NewUnion(graphql.UnionConfig{
	Name: "Payload",
	Types: []*graphql.Object{
		documentType,
		reportType,
	},
	ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
		// Resolver для union типа, в котором определяется с payload-ом какого типа в данный момент имеем дело
		// на основании текущего значения p.Value.
		switch p.Value.(type) {
		case Document:
			return documentType
		case Report:
			return reportType
		default:
			panic(errors.New("unable to find appropriate type"))
		}
	},
})

var eventType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Event",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"payload": &graphql.Field{
			Type: unionType,
		},
	},
})

func main() {
	fmt.Println("Starting application...")

	// Пример тестовых данных. Допустим, что они получены из базы.
	var data = []Event{
		{
			ID:   "1",
			Name: FileUploadedEvent,
			Type: "direct",
			Payload: Document{
				ID:    "1",
				Title: "This is a document",
			},
		},
		{
			ID:   "2",
			Name: ReportCreatedEvent,
			Type: "group",
			Payload: Report{
				ID:   "2",
				Name: "This is a report",
			},
		},
	}

	var rootQuery = graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"events": &graphql.Field{
				Type: graphql.NewList(eventType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return data, nil
				},
			},
		},
	})
	var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)

		var body RequestBody
		decoder.Decode(&body)
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: body.Query,
		})
		json.NewEncoder(w).Encode(result)
	})
	http.ListenAndServe(fmt.Sprintf("%s:%s", ServerHost, ServerPort), nil)
}
