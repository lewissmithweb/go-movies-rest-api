package dbrepo

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}

const dbTimeout = time.Second * 3

func (m *PostgresDBRepo) AllMovies(genre ...int) ([]*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	where := ""
	if len(genre) > 0 {
		where = fmt.Sprintf("where id in (select movie_id from genres where genre_id = %d)", genre[0])
	}

	query := fmt.Sprintf(`
		select
			id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		from
		    movies %s
		order by
		    title
	`, where)

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*models.Movie

	for rows.Next() {
		var movie models.Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.ReleaseDate,
			&movie.Runtime,
			&movie.MPAARating,
			&movie.Description,
			&movie.Image,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	return movies, nil
}

func (m *PostgresDBRepo) MovieByIDForEdit(id int) (*models.Movie, []*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		from
		    movies
		where id = $1
	`

	var movie models.Movie
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.Runtime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil, nil, err
	}

	query = `select g.id, g.genre from movies_genres mg left join genres g on g.id = mg.genre_id where mg.movie_id = $1 order by g.genre`
	rows, err := m.DB.QueryContext(ctx, query, movie.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	var genresArray []int
	for rows.Next() {
		var genre models.Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
		)
		if err != nil {
			return nil, nil, err
		}

		genres = append(genres, &genre)
		genresArray = append(genresArray, genre.ID)
	}

	movie.Genres = genres
	movie.GenresArray = genresArray

	var allGenres []*models.Genre

	query = `select id, genre from genres order by genre`
	gRows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer gRows.Close()

	for gRows.Next() {
		var genre models.Genre
		err := gRows.Scan(
			&genre.ID,
			&genre.Name,
		)
		if err != nil {
			return nil, nil, err
		}

		allGenres = append(allGenres, &genre)
	}

	return &movie, allGenres, nil
}

func (m *PostgresDBRepo) MovieByID(id int) (*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, title, release_date, runtime, mpaa_rating, description, coalesce(image, ''), created_at, updated_at
		from
		    movies
		where id = $1
	`

	var movie models.Movie
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.Runtime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	query = `select g.id, g.genre from movies_genres mg left join genres g on g.id = mg.genre_id where mg.movie_id = $1 order by g.genre`
	rows, err := m.DB.QueryContext(ctx, query, movie.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	for rows.Next() {
		var genre models.Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &genre)
	}

	movie.Genres = genres
	return &movie, nil
}

func (m *PostgresDBRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, created_at, updated_at from users where email = $1`

	var user models.User
	row := m.DB.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *PostgresDBRepo) GetUserByID(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, first_name, last_name, password, created_at, updated_at from users where id = $1`

	var user models.User
	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *PostgresDBRepo) AllGenres() ([]*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, genre, created_at, updated_at from genres`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	for rows.Next() {
		var genre models.Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.CreatedAt,
			&genre.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		genres = append(genres, &genre)
	}

	return genres, nil
}

func (m *PostgresDBRepo) InsertMovie(movie models.Movie) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `insert into movies (title, description, release_date, runtime, mpaa_rating, created_at, updated_at, image) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`
	var newId int

	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		movie.Title,
		movie.Description,
		movie.ReleaseDate,
		movie.Runtime,
		movie.MPAARating,
		movie.CreatedAt,
		movie.UpdatedAt,
		movie.Image,
	).Scan(&newId)
	if err != nil {
		return 0, err
	}

	return newId, nil
}

func (m *PostgresDBRepo) UpdateMovie(movie models.Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `update movies set title = $1, description = $2, release_date = $3, runtime = $4, mpaa_rating = $5, updated_at = $6, image = $7 where id = $8`
	_, err := m.DB.ExecContext(ctx, stmt,
		movie.Title,
		movie.Description,
		movie.ReleaseDate,
		movie.Runtime,
		movie.MPAARating,
		movie.UpdatedAt,
		movie.Image,
		movie.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *PostgresDBRepo) UpdateMovieGenres(id int, genreIDs []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from movies_genres where id = $1`
	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	for _, g := range genreIDs {
		stmt := `insert into movies_genres (id, movie_id) values ($1, $2)`
		_, err := m.DB.ExecContext(ctx, stmt, g, id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *PostgresDBRepo) DeleteMovie(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from movies where id = $1`
	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}
