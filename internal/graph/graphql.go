package graph

import (
	"backend/internal/models"
	"errors"
	"github.com/graphql-go/graphql"
	"strings"
)

type Graph struct {
	Movies      []*models.Movie
	QueryString string
	Config      graphql.SchemaConfig
	fields      graphql.Fields
	movieType   *graphql.Object
}

func New(movies []*models.Movie) *Graph {
	var movieType = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Movie",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.Int,
				},
				"title": &graphql.Field{
					Type: graphql.String,
				},
				"description": &graphql.Field{
					Type: graphql.String,
				},
				"release_date": &graphql.Field{
					Type: graphql.DateTime,
				},
				"runtime": &graphql.Field{
					Type: graphql.Int,
				},
				"mpaa_rating": &graphql.Field{
					Type: graphql.String,
				},
				"created_at": &graphql.Field{
					Type: graphql.DateTime,
				},
				"updated_at": &graphql.Field{
					Type: graphql.DateTime,
				},
				"image": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)

	var fields = graphql.Fields{
		"list": &graphql.Field{
			Type:        graphql.NewList(movieType),
			Description: "List of movies",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return movies, nil
			},
		},
		"search": &graphql.Field{
			Type:        graphql.NewList(movieType),
			Description: "Search movies",
			Args: graphql.FieldConfigArgument{
				"title_contains": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var movieList []*models.Movie
				search, ok := p.Args["title_contains"].(string)
				if ok {
					for _, movie := range movieList {
						if strings.Contains(strings.ToLower(movie.Title), strings.ToLower(search)) {
							movieList = append(movieList, movie)
						}
					}
				}
				return movieList, nil
			},
		},
		"get": &graphql.Field{
			Type:        graphql.NewList(movieType),
			Description: "Get movie by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id, ok := p.Args["id"].(int)
				if ok {
					for _, movie := range movies {
						if id == movie.ID {
							return movie, nil
						}
					}
				}
				return nil, nil
			},
		},
	}

	return &Graph{
		Movies:    movies,
		fields:    fields,
		movieType: movieType,
	}
}

func (g *Graph) Query() (*graphql.Result, error) {
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: g.fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, err
	}

	params := graphql.Params{Schema: schema, RequestString: g.QueryString}
	response := graphql.Do(params)
	if len(response.Errors) > 0 {
		return nil, errors.New(response.Errors[0].Message)
	}

	return response, nil
}
