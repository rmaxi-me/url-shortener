# url-shortener
Lightweight URL Shortener written in Go with PostgreSQL

## Get started
### Environment Variables

| Name                | Description                                                 |
|---------------------|-------------------------------------------------------------|
| `URL_ID_LENGTH`     | Positive integer for how long shortened links IDs should be |
| `SERVER_PORT`       | Backend server port                                         |
| `POSTGRES_DB`       | Database name                                               |
| `POSTGRES_USER`     | Database username                                           |
| `POSTGRES_PASSWORD` | Database password                                           |
| `POSTGRES_IP`       | Database IP address                                         |
| `POSTGRES_PORT`     | Database port                                               |

### Docker Compose
Set up a `.env` with your desired values (see [.env.example](.env.example) for an example file), then you can run `docker compose up` to get the database running, and `docker compose down` to shut it down.

### Running the backend
To start the backend, make sure your `.env` is set up, and run `go run rmaxi-me/url-shortner/src/cmd`.

### Running the tests
To run the tests, use `go test -v ./tests/...`.

## Planned features
- [ ] Dockerfile for deployment
- [ ] Optional time to live for links.
- [ ] Have a way for a user to know how many times their links have been accessed.
  - [x] Backend already stores and increments this, we just need a stat endpoint.
- [ ] "Last visited at" property.

## AI Disclosure
AI was used only to generate tests, and the code was double-checked. Everything else is handwritten by me.
