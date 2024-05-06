CREATE TABLE Movie (
    movie_id INT NOT NULL AUTO_INCREMENT,
    title VARCHAR(500) NOT NULL,
    movie_url VARCHAR(2083) NOT NULL,
    poster_url VARCHAR(2083) NOT NULL,
    PRIMARY KEY (movie_id)
);