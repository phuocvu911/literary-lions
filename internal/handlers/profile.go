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
    LikedPosts []*FullPost
}

type PageData struct {
    User *models.User
    Content Content
    PostCount int 
    CommentCount int
    LikedCount int
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
    likedPosts := app.GetUserLikedPosts(user.ID)
    comments := app.GetUserComments(user.ID)

    data := PageData{
        User: &user,
        Content: Content{
            Posts: posts, 
            Comments: comments,
            LikedPosts: likedPosts,
        },
        PostCount: len(posts), 
        CommentCount: len(comments),
        LikedCount: len(likedPosts),
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
        fmt.Println("Database error:", err)
        return
    } 
    
    if len(imageData) == 0 {
        fmt.Println("No profile image, serving default")
        http.ServeFile(w, r, "internal/web/static/profile.jpeg")
        return
    }

    
    w.Header().Set("Content-Type", "image/jpeg")
    w.Write(imageData)
}

func (app *App) ProfileDeleteHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)        
        fmt.Println("here")
        return
    }

    if app.currentUser(r) == nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    user := app.currentUser(r)

    sql := `DELETE FROM users WHERE id = ?`

    _, err := app.db.Exec(sql, user.ID)
    if err != nil {
        return
    }
    
    http.Redirect(w, r, "/", http.StatusSeeOther)
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

func (app *App) GetUserLikedPosts(userId int64) []*FullPost{
    queryLikedPosts := `
        SELECT
            posts.id AS post_id,
            posts.title AS post_title,
            posts.content AS post_content,
            posts.created_at AS post_created_at,
            post_reactions.value AS reaction
        FROM posts
        JOIN post_reactions
            ON post_reactions.post_id = posts.id
        WHERE post_reactions.user_id = ?;
    `

    rows, err := app.db.Query(queryLikedPosts, userId)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    defer rows.Close()

    posts := []*FullPost{}

    for rows.Next() {
        p := &FullPost{}

        if err := rows.Scan(
            &p.ID,
            &p.Title, 
            &p.Content, 
            &p.CreatedAt,         
            &p.CurrentReaction,  
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