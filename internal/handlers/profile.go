package handlers

import (
	"net/http"
    "net/mail"
    "lions/internal/models"
    "lions/internal/database"
    "io"
    "fmt"	
    "strings"
)

type Content struct {
    Posts []*models.Post
    Comments []*models.Comment
}

type PageData struct {
    User *models.User
    Content Content
    PostCount int 
    CommentCount int
}

type EditPageData struct {
    User *models.User
    UserName string
    Email    string
    Error string
}

func (app *App) ProfileHandler(w http.ResponseWriter, r *http.Request) {
    
    if app.currentUser(r) == nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    ID := app.currentUser(r).ID
    getUser := `SELECT * FROM users WHERE id = ?`

    row := app.db.QueryRow(getUser, ID)
    user := models.User{}

    err := row.Scan(&user.ID,&user.Email,&user.Username, &user.PasswordHash, &user.CreatedAt, &user.ProfileImage)
    if err != nil {
        fmt.Println(err)
        return 
    }

    posts := app.GetUserPosts(user.ID)
    comments := app.GetUserComments(user.ID)

    data := PageData{
        User: &user,
        Content: Content{
            Posts: posts, 
            Comments: comments,
        },
        PostCount: len(posts), 
        CommentCount: len(comments),
    }

    app.render(w, "profile.html", data)
}

func (app *App) ProfileUploadHandler(w http.ResponseWriter, r *http.Request) {
    if app.currentUser(r) == nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    file, header, err := r.FormFile("profile_image")
    if err != nil {
        http.Error(w, "Could not get image", http.StatusBadRequest)
        return
    }
    defer file.Close()

    fmt.Println("Uploading:", header.Filename)

    imageData, err := io.ReadAll(file)
    if err != nil {
        http.Error(w, "Could not read image", http.StatusInternalServerError)
        return
    }

    ID := app.currentUser(r).ID

    _, err = app.db.Exec(`
        UPDATE users
        SET profile_image = ?
        WHERE id = ?
    `, imageData, ID)

    if err != nil {
        http.Error(w, "Could not save image", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (app *App) ProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user := app.currentUser(r)

	if user == nil {
		http.Error(w, "Not signed in", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Show the form
		data := EditPageData{
			User:     user,
			UserName: user.Username,
			Email:    user.Email,
		}

		app.render(w, "edit-profile.html", data)

	case http.MethodPost:
		// Process the form
		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(r.FormValue("email"))
        currentPassword := r.FormValue("current-password")
		newPassword := r.FormValue("new-password")

        //fail scenarios closure
        fail := func(msg string) {
            w.WriteHeader(http.StatusUnprocessableEntity)
            app.render(w, "edit-profile.html", EditPageData{
                User:     user,
                UserName: username,
                Email:    email,
                Error:    msg,
            })
        }

        //validate input
        if _, err := mail.ParseAddress(email); err != nil || len(email) > maxCredentials {
            fail("Please enter a valid email address.")
            return
        }
        if len(username) < 3 || len(username) > maxCredentials || strings.ContainsAny(username, " \t\v\n@") {
            fail("Username must be at least 3 characters, with no spaces or '@'.")
            return
        }

        //check email and password match
        userCheck, errtwo := database.UserByEmail(app.db, email)
        if errtwo == nil {
            if isRightPassword := app.hasher.Verify(currentPassword, userCheck.PasswordHash); !isRightPassword {
                fail("Email or password is incorrect.")
                return
            }
        }
        
        if newPassword != "" {
            if len(newPassword) < 8 || len(newPassword) > maxCredentials {
                fail("Password must be at least 8 characters.")
                return
            }
        }    

		var err error

		if strings.TrimSpace(newPassword) == "" {
			_, err = app.db.Exec(`
				UPDATE users
				SET username = ?, email = ?
				WHERE id = ?
			`, username, email, user.ID)
		} else {
			hash, hashErr := app.hasher.Hash(newPassword)
			if hashErr != nil {
				app.serverError(w, hashErr)
				return
			}

			_, err = app.db.Exec(`
				UPDATE users
				SET username = ?, email = ?, password_hash = ?
				WHERE id = ?
			`, username, email, hash, user.ID)
		}

		if err != nil {
			fmt.Println(err)
			http.Error(w, "Could not update the profile", http.StatusInternalServerError)
			return
		}

		// Redirect once done
		http.Redirect(w, r, "/profile", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) ProfileImageHandler(w http.ResponseWriter, r *http.Request) {
    if app.currentUser(r) == nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }    

    ID := app.currentUser(r).ID
    var imageData []byte

    err := app.db.QueryRow(`
        SELECT profile_image
        FROM users
        WHERE id = ?
    `, ID).Scan(&imageData)

    if err != nil {
        http.Error(w, "Image not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "image/jpeg")
    w.Write(imageData)
}

//helper
func (app *App) GetUserPosts(userId int64) []*models.Post{
    queryPosts := `SELECT * FROM posts WHERE user_id = ?`

    rows, err := app.db.Query(queryPosts, userId)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    defer rows.Close()

    posts := []*models.Post{}

    for rows.Next() {
        p := &models.Post{}

        if err := rows.Scan(&p.ID, 
            &p.ID,
            &p.Title, 
            &p.Content, 
            &p.CreatedAt,          
        ); err != nil {
            fmt.Println(err)
            return nil
        }

        posts = append(posts, p)
    }

    if err = rows.Err(); err != nil {
        fmt.Println(err)
        return nil 
    }

    return posts
}

func (app *App) GetUserComments(userId int64) []*models.Comment{
    queryComments := `SELECT * FROM comments WHERE user_id = ?`

    rows, err := app.db.Query(queryComments, userId)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    defer rows.Close()

    comments := []*models.Comment{}

    for rows.Next() {
        c := &models.Comment{}

        if err := rows.Scan(&c.ID, 
            &c.ID,
            &c.ID,
            &c.Content, 
            &c.CreatedAt,           
        ); err != nil {
            fmt.Println(err)
            return nil
        }

        comments = append(comments, c)
    }

    if err = rows.Err(); err != nil {
        fmt.Println(err)
        return nil 
    }

    return comments
}