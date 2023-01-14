package dbrepo

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type PostgresDbRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 3

func (m *PostgresDbRepo) Connection() *sql.DB {
	return m.DB
}

func (m *PostgresDbRepo) AllMovies(genre ...int) ([]*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	where := ""
	if len(genre) > 0 {
		where = fmt.Sprintf("where id in (select movie_id from movies_genres where genre_id=%d)", genre[0])
	}

	query := fmt.Sprintf(`
	SELECT 
		id, title, release_date, runtime,
		mpaa_rating, description, coalesce(image, ''),
		created_at, updated_at
	FROM
		movies %s
	ORDER BY
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
			&movie.RunTime,
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

func (m *PostgresDbRepo) OneMovie(id int) (*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT 
		id, title, release_date, runtime,
		mpaa_rating, description, coalesce(image, ''),
		created_at, updated_at
	FROM
		movies
	WHERE
		id = $1
	`

	var movie models.Movie
	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	query = `select g.id, g.genre from movies_genres mg left join genres g on (mg.genre_id=g.id)
			where mg.movie_id=$1 order by g.genre
			`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
		)

		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}
	movie.Genres = genres
	return &movie, nil
}

func (m *PostgresDbRepo) OneMovieForEdit(id int) (*models.Movie, []*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT 
		id, title, release_date, runtime,
		mpaa_rating, description, coalesce(image, ''),
		created_at, updated_at
	FROM
		movies
	WHERE
		id = $1
	`

	var movie models.Movie
	row := m.DB.QueryRowContext(ctx, query, id)
	log.Println("This places has passed 1")
	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.ReleaseDate,
		&movie.RunTime,
		&movie.MPAARating,
		&movie.Description,
		&movie.Image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	log.Println("This places has passed 2")
	if err != nil {
		return nil, nil, err
	}
	log.Println("This places has passed 3")

	query = `select g.id, g.genre from movies_genres mg left join genres g on (mg.genre_id=g.id)
			where mg.movie_id = $1 order by g.genre
			`
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}
	log.Println("This places has passed 4")
	defer rows.Close()

	var genres []*models.Genre
	var genresArray []int
	for rows.Next() {
		var g models.Genre
		err := rows.Scan(
			&g.ID,
			&g.Genre,
		)

		if err != nil {
			return nil, nil, err
		}

		genres = append(genres, &g)
		genresArray = append(genresArray, g.ID)
	}
	log.Println("This places has passed 5")
	movie.Genres = genres
	movie.GenresArray = genresArray
	var allGenres []*models.Genre
	query = `select id, genre from genres order by genre`
	grows, err := m.DB.QueryContext(ctx, query)
	log.Println("This places has passed 6")
	if err != nil {
		return nil, nil, err
	}
	log.Println("This places has passed 7")
	defer grows.Close()
	for grows.Next() {
		var g models.Genre
		err := grows.Scan(
			&g.ID,
			&g.Genre,
		)
		if err != nil {
			return nil, nil, err
		}

		allGenres = append(allGenres, &g)
	}
	log.Println("This places has passed 8")
	return &movie, allGenres, err
}

func (m *PostgresDbRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, email, first_name, last_name, password,
		created_at, updated_at from users where email = $1`

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

func (m *PostgresDbRepo) GetUserById(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, email, first_name, last_name, password,
		created_at, updated_at from users where id = $1`

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

func (m *PostgresDbRepo) AllGenres() ([]*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, genre, created_at, updated_at from genres order by genre`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var genres []*models.Genre

	for rows.Next() {
		var g models.Genre

		err := rows.Scan(
			&g.ID,
			&g.Genre,
			&g.CreatedAt,
			&g.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		genres = append(genres, &g)
	}

	return genres, nil

}

func (m *PostgresDbRepo) InsertMovie(movie models.Movie) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `insert into movies 
				(
					title,
					description,
					release_date,
					runtime,
					mpaa_rating,
					created_at,
					updated_at,
					image
				)
				values (
					$1, $2, $3, $4, $5, $6, $7, $8
				)
				returning id
			`
	var newID int
	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		movie.Title,
		movie.Description,
		movie.ReleaseDate,
		movie.RunTime,
		movie.MPAARating,
		movie.CreatedAt,
		movie.UpdatedAt,
		movie.Image,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *PostgresDbRepo) UpdateMovie(movie models.Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	update movies set 
		title=$1, description=$2, release_date=$3, runtime=$4,
		mpaa_rating=$5, updated_at=$6, image=$7
	where id=$8
	`
	_, err := m.DB.ExecContext(
		ctx,
		query,
		movie.Title,
		movie.Description,
		movie.ReleaseDate,
		movie.RunTime,
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

func (m *PostgresDbRepo) UpdateMovieGenres(id int, genreIDs []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from movies_genres where movie_id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	for _, n := range genreIDs {
		stmt := `insert into movies_genres (movie_id, genre_id) values ($1, $2)`
		_, err := m.DB.ExecContext(ctx, stmt, id, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *PostgresDbRepo) DeleteMovieById(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `delete from movies where id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}
