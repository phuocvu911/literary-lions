package handlers

import (
	"net/http"
    "net/mail"
    "lions/internal/models"
    "database/sql"
    "errors"
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

    err := row.Scan(&user.ID,&user.Email,&user.Username, &user.PasswordHash, &user.CreatedAt, &user.ProfileImage, &user.ProfileImageVersion)
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
    user := app.currentUser(r)

    if user == nil {
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

    _, err = app.db.Exec(`
        UPDATE users
        SET profile_image = ?,
            image_version = image_version + 1
        WHERE id = ?
    `, imageData, user.ID)

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
        username := strings.TrimSpace(r.FormValue("username"))
        email := strings.TrimSpace(r.FormValue("email"))
        currentPassword := r.FormValue("current-password")
        newPassword := r.FormValue("new-password")

        if err := app.ValidateInput(user, username, email, currentPassword, newPassword); err != nil {
            if errors.Is(err, sql.ErrNoRows) {
                app.serverError(w, err)
                return
            }
            
            app.render(w, "edit-profile.html", EditPageData{
                User:     user,
                UserName: username,
                Email:    email,
                Error:    err.Error(),
            })
            return
        }

        var err error

        if newPassword == "" {
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
            app.serverError(w, err)
            return
        }

        http.Redirect(w, r, "/profile", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) ProfileImageHandler(w http.ResponseWriter, r *http.Request) {
    user := app.currentUser(r)

    if user == nil {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }    

    var imageData []byte

    err := app.db.QueryRow(`
        SELECT profile_image
        FROM users
        WHERE id = ?
    `, user.ID).Scan(&imageData)

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

func (app *App) ValidateInput(user *models.User,username, email, currentPassword, newPassword string) error {
    // Username
    if len(username) < 3 || len(username) > maxCredentials || strings.ContainsAny(username, " \t\v\n@") {
        return errors.New("Username must be at least 3 characters, with no spaces or '@'.")
    }

    if username != user.Username {
        var existing string

        err := app.db.QueryRow(`SELECT username FROM users WHERE username = ? AND id != ?`, username, user.ID).Scan(&existing)

        if err == nil {
            return errors.New("The username is already in use. Please choose a new one.")
        }

        if !errors.Is(err, sql.ErrNoRows) {
            return err
        }
    }

    // Email
    if email == "" || len(email) > maxCredentials {
        return errors.New("Please enter a valid email address.")
    }

    parsed, err := mail.ParseAddress(email)
    if err != nil || parsed.Address != email {
        return errors.New("Please enter a valid email address.")
    }

    if email != user.Email {
        var existing string

        err := app.db.QueryRow(`SELECT email FROM users WHERE email = ? AND id != ?`,email, user.ID).Scan(&existing)

        if err == nil {
            return errors.New("The email is already in use. Please choose a new one.")
        }

        if !errors.Is(err, sql.ErrNoRows) {
            return err
        }
    }

    // Current password
    if currentPassword == "" {
        return errors.New("Please enter your current password.")
    }

    var passwordHash string
    err = app.db.QueryRow(`SELECT password_hash FROM users WHERE id = ?`, user.ID).Scan(&passwordHash)

    if err != nil {
        return fmt.Errorf("failed to retrieve password hash: %w", err)
    }

    if !app.hasher.Verify(currentPassword, passwordHash) {
        return errors.New("Current password is incorrect.")
    }

    // New password
    if newPassword != "" {
        if newPassword == currentPassword {
            return errors.New("New password should be different from your old password.")
        }

        if len(newPassword) < 8 || len(newPassword) > maxCredentials {
            return errors.New("Password must be at least 8 characters.")
        }
    }

    return nil
}
