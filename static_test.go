package static

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// nolint:unparam
func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequestWithContext(context.Background(), method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestEmptyDirectory(t *testing.T) {
	// SETUP file
	testRoot, _ := os.Getwd()
	f, err := ioutil.TempFile(testRoot, "")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("Gin Web Framework")
	f.Close()

	dir, filename := filepath.Split(f.Name())

	router := gin.New()
	router.Use(ServeRoot("/", dir))
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "index")
	})
	router.GET("/a", func(c *gin.Context) {
		c.String(http.StatusOK, "a")
	})
	router.GET("/"+filename, func(c *gin.Context) {
		c.String(http.StatusOK, "this is not printed")
	})
	w := performRequest(router, "GET", "/")
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "index")

	w = performRequest(router, "GET", "/"+filename)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Gin Web Framework")

	w = performRequest(router, "GET", "/"+filename+"a")
	assert.Equal(t, w.Code, 404)

	w = performRequest(router, "GET", "/a")
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "a")

	router2 := gin.New()
	router2.Use(ServeRoot("/static", dir))
	router2.GET("/"+filename, func(c *gin.Context) {
		c.String(http.StatusOK, "this is printed")
	})

	w = performRequest(router2, "GET", "/")
	assert.Equal(t, w.Code, 404)

	w = performRequest(router2, "GET", "/static")
	assert.Equal(t, w.Code, 404)
	router2.GET("/static", func(c *gin.Context) {
		c.String(http.StatusOK, "index")
	})

	w = performRequest(router2, "GET", "/static")
	assert.Equal(t, w.Code, http.StatusOK)

	w = performRequest(router2, "GET", "/"+filename)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "this is printed")

	w = performRequest(router2, "GET", "/static/"+filename)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Gin Web Framework")
}

func TestIndex(t *testing.T) {
	// SETUP file
	testRoot, _ := os.Getwd()
	f, err := os.Create(path.Join(testRoot, "index.html"))
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("index")
	f.Close()

	dir, filename := filepath.Split(f.Name())

	router := gin.New()
	router.Use(ServeRoot("/", dir))

	w := performRequest(router, "GET", "/"+filename)
	assert.Equal(t, w.Code, 301)

	w = performRequest(router, "GET", "/")
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "index")
}

func TestListIndex(t *testing.T) {
	// SETUP file
	testRoot, _ := os.Getwd()
	f, err := ioutil.TempFile(testRoot, "")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("Gin Web Framework")
	f.Close()

	dir, filename := filepath.Split(f.Name())
	router := gin.New()
	router.Use(Serve("/", LocalFile(dir, true)))

	w := performRequest(router, "GET", "/"+filename)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), "Gin Web Framework")

	w = performRequest(router, "GET", "/")
	assert.Contains(t, w.Body.String(), `<a href="`+filename)
}
