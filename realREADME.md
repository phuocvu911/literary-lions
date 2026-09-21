# Literary Lions Forum 
 
A Go web forum for a book club: SQLite-backed, cookie/UUID sessions, categorized
posts, like/dislike, filtering, and Docker deployment.

## Requirements

- Go 1.27+
- [Docker](https://docs.docker.com/get-started/get-docker/) (for containerized deployment)
- GCC (required if you use Windows machine, install [MinGW](https://www.mingw-w64.org/) for C compiler support)

## How To Run
Clone the repository and run the following commands:

```
git clone https://gitea.kood.tech/hoangphuocvu/forum
cd forum
```

Docker deployment:

...

## How It Works
### ERD
![ERD Diagram](/erd.png)
...

## Extras

### Password Hashing

The forum uses sha256 for password hashing. When a user registers, their password is hashed using the sha256 algorithm before being stored in the database. This ensures that even if the database is compromised, the actual passwords remain secure. Powered by the `crypto/sha256` package in Go.

### Session Management

When a user logs in, a unique session ID (uuid) is generated and stored in a cookie on the user's browser. This session ID is also stored in the database, allowing the server to identify the user on subsequent requests. The session management system ensures that users remain logged in across different pages of the forum. Powered by the `uuid` package in Go since Go 1.27.0.

## Bonus Features

### Bcrypt Password Hashing

Hashing passwords with bcrypt is more secure than sha256. Bcrypt automatically handles salting and is designed to be computationally expensive, making it resistant to brute-force attacks. This feature is implemented by using the `golang.org/x/crypto/bcrypt` package.

It solves the problem of [rainbow table](https://www.huntress.com/cybersecurity-101/topic/rainbow-table-defined), which are precomputed lookup tables for matching the stolen hash (because 2 indentical passwords would produce the same hash). Since bcrypt includes a salt, it makes rainbow table attacks ineffective.

You can run it with ` --bcrypt` flag to enable bcrypt password hashing.

```
go run main.go --bcrypt
```

### 404 Page
The 404 page is a simple soft landing page that shows the user that the page they are looking for does not exist if they navigate to a non-existent URL. The user can click the button to go back to the home page.

### Reset Password
Users can reset their passwords if they forgot it. The reset password link will be sent to their registered email address. Only registered email can receive the reset password link, but the error message will be generic to prevent the disclosure of registered email addresses. The link is valid for 30 minutes. Powered by `net/smtp` package.

You can try to register an account with your email to see the reset password feature in action. If you don't receive the reset link, please check your spam folder.

### Rate Limiter

The rate limiter powered by `github.com/phuocvu911/ratelimiter` package is used to prevent abuse of the forum by limiting the number of requests a user can make in a given time frame. This helps to protect the server from being overwhelmed by too many requests and ensures fair usage for all users.

### Search Feature

Xinyu

### Upload Images

Aman

### User Profile Page

Aman