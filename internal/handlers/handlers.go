package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
    "lions/internal/models"
    "io"
    "fmt"
)

type App struct {
	db        *sql.DB
	templates map[string]*template.Template
}

type Content struct {
    Posts []*models.Post
    Comments []*models.Comment
}

//mock struct
type PageData struct {
    User *models.User
    Content Content
    PostCount int 
    CommentCount int
}

func NewApp(db *sql.DB, templates map[string]*template.Template) *App {
	return &App{db: db, templates: templates}
}

func (app *App) ProfileHandler(w http.ResponseWriter, r *http.Request) {

    getUser := `SELECT * FROM users WHERE id = ?`
    row := app.db.QueryRow(getUser, 1)
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
        SET profile_image = ?
        WHERE id = 1
    `, imageData)

    if err != nil {
        http.Error(w, "Could not save image", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/profile", http.StatusSeeOther)
}


func (app *App) ProfileImageHandler(w http.ResponseWriter, r *http.Request) {
    var imageData []byte

    err := app.db.QueryRow(`
        SELECT profile_image
        FROM users
        WHERE id = 1
    `).Scan(&imageData)

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