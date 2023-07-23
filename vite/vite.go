package vite

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

type Manifest struct {
	entries           map[string]Entry
	url               string
	publicPath        string
	manifestDirectory string
	manifestPath      string
	hotProxyURL       string
}

type Entry struct {
	File string `json:"file"`
	// Src  string `json:"src"`

	Assets []string `json:"assets"`
	CSS    []string `json:"css"`

	IsEntry bool `json:"isEntry"`
}

// New function.
func New(url, publicPath, manifestDirectory, hotProxyURL string) (*Manifest, error) {
	entries := make(map[string]Entry)

	m := &Manifest{
		url:         url,
		publicPath:  publicPath,
		hotProxyURL: hotProxyURL,
		entries:     entries,
	}
	m.manifestDirectory = m.pathPrefix(manifestDirectory)
	m.manifestPath = m.getManifestPath(m.publicPath + m.manifestDirectory)
	content, err := os.ReadFile(m.manifestPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(content, &m.entries)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// func (m *Manifest) manifestPath(manifestDirectory string) string {
// 	return manifestDirectory + "/manifest.json"
// }

func (m *Manifest) pathPrefix(path string) string {
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}

// Hash function.
func (m *Manifest) Hash(manifestDirectory string) (string, error) {
	// manifestPath := m.manifestPath

	_, err := os.Stat(m.manifestPath)
	if os.IsNotExist(err) {
		return "", ErrManifestNotExist
	}
	file, err := os.Open(m.manifestPath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	log.Println("mix file")
	log.Println(m.hashFromFile(file))
	return m.hashFromFile(file)
}

// HashFromFS function.
func (m *Manifest) HashFromFS(manifestDirectory string, staticFS fs.FS) (string, error) {
	file, err := staticFS.Open(m.getManifestPath(manifestDirectory))
	// file, err := staticFS.Open(strings.TrimPrefix(m.manifestPath, "/"))
	if err != nil {
		return "", err
	}

	defer file.Close()

	return m.hashFromFile(file)
}

func (m *Manifest) hashFromFile(file io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
func (m *Manifest) LoadViteEmbed(entry string) (string, error) {
	info, err := os.Stat(m.publicPath + "/hot")
	log.Printf("error %v info %v", err, info)
	if err == nil {
		if m.hotProxyURL != "" {
			return m.hotProxyURL + entry, nil
		}

		content, err := os.ReadFile(m.publicPath + "/hot")
		if err != nil {
			return "", err
		}
		log.Printf("hot-content %v", content)
		url := strings.TrimSpace(string(content))
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			url = url[strings.Index(url, ":")+1:]
		} else {
			url = "//localhost:8080"
		}

		return m.GetScript(url + m.pathPrefix(entry)), nil
	}
	if _, ok := m.entries[entry]; !ok {
		log.Printf("vite: unable to locate vite file: %v", entry)
		return "", fmt.Errorf("vite: unable to locate vite file: %v", entry)
	}
	return m.GetScript(m.buildPath(m.entries[entry].File)) + m.jsPreloadImports(entry) + m.GetStyles(entry), nil
}
func (m *Manifest) GetAsset(path string) string {
	return fmt.Sprintf(`<script type="module" src="%s"></script>`, m.buildPath(m.entries[path].File))
}
func (m *Manifest) GetScript(path string) string {
	return fmt.Sprintf(`<script type="module" src="%s"></script>`, path)
}
func (m *Manifest) jsPreloadImports(name string) string {
	return ""
}
func (m *Manifest) GetJS(name string) string {
	return m.buildPath(m.entries[name].File)
}

func (m *Manifest) GetStyles(name string) string {
	entry := m.entries[name]
	var links string
	for _, css := range entry.CSS {
		links += fmt.Sprintf(`<link rel="stylesheet" href="%s">`, m.buildPath(css))
	}
	return links
}

func (m *Manifest) getManifestPath(path string) string {
	return path + "/manifest.json"
}
func (m *Manifest) buildPath(path string) string {
	return m.url + m.manifestDirectory + m.pathPrefix(path)
}
