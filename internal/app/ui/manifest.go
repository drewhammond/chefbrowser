package ui

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

type Vite struct {
	HTMLTags string
	manifest map[string]viteItem
	cfg      ViteConfig
}

type ViteConfig struct {
	Base        string
	Environment string
	Manifest    []byte
	FS          fs.FS
}

type viteItem struct {
	File    string   `json:"file"`
	Src     string   `json:"src"`
	IsEntry bool     `json:"isEntry"`
	CSS     []string `json:"css"`
	Assets  []string `json:"assets"`
}
type ViteManifest struct {
	Files map[string]viteItem
}

func NewVite(cfg ViteConfig) (*Vite, error) {
	v := Vite{}
	v.cfg = cfg
	err := v.generateTags()
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (v *Vite) generateTags() error {
	if v.cfg.Environment == "production" {
		err := v.parseManifest()
		if err != nil {
			return err
		}
	} else {
		v.HTMLTags = `<script type="module" src="http://localhost:5173/@vite/client"></script>
<script type="module" src="http://localhost:5173/main.js"></script>`
	}

	return nil
}

func (v *Vite) parseManifest() error {
	err := json.Unmarshal(v.cfg.Manifest, &v.manifest)
	if err != nil {
		return err
	}

	var tmpl string
	// for now, we're only interested in index.js; we may eventually use other functionality in Vite
	for _, y := range v.manifest {
		if !y.IsEntry {
			continue
		}

		for _, cssFile := range y.CSS {
			tmpl += fmt.Sprintf(`<link rel="stylesheet" href="%s" />`, v.cfg.Base+"/"+cssFile)
		}

		tmpl += fmt.Sprintf(`<script type="module" src="%s"></script>`, v.cfg.Base+"/"+y.File)

	}

	v.HTMLTags = tmpl

	return nil
}
