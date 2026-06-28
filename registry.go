package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type ManifestResponse struct {
	Layers []struct {
		Digest string `json:"digest"`
	} `json:"layers"`
	
	Manifests []struct {
		Digest   string `json:"digest"`
		Platform struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
		} `json:"platform"`
	} `json:"manifests"`
}

func pullImage(image string) {
	fmt.Printf("Pulling image metadata for: %s\n", image)

	tokenUrl := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull", image)
	resp, err := http.Get(tokenUrl)
	must(err)
	defer resp.Body.Close()

	var tokenData TokenResponse
	must(json.NewDecoder(resp.Body).Decode(&tokenData))
	fmt.Println("Successfully authenticated with Docker Hub!")

	manifest := fetchManifest(image, "latest", tokenData.Token)

	if len(manifest.Layers) == 0 && len(manifest.Manifests) > 0 {
		fmt.Println("Multi-architecture image detected. Finding the Linux/amd64 version...")
		for _, m := range manifest.Manifests {
			if m.Platform.Architecture == "amd64" && m.Platform.OS == "linux" {
				manifest = fetchManifest(image, m.Digest, tokenData.Token)
				break
			}
		}
	}

	fmt.Printf("Found %d layers for %s!\n", len(manifest.Layers), image)
	
	layerDir := filepath.Join(os.Getenv("HOME"), ".gobox", "layers", image)
	os.MkdirAll(layerDir, 0755)

	for i, layer := range manifest.Layers {
		digest := layer.Digest
		fmt.Printf("   Downloading Layer %d: %s\n", i+1, digest[:15]+"...")
		downloadLayer(image, digest, tokenData.Token, layerDir)
	}
}

func fetchManifest(image, reference, token string) ManifestResponse {
	manifestUrl := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/manifests/%s", image, reference)
	req, err := http.NewRequest("GET", manifestUrl, nil)
	must(err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.oci.image.index.v1+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	must(err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Failed to get manifest. HTTP Status: %d\nBody: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var manifest ManifestResponse
	must(json.NewDecoder(resp.Body).Decode(&manifest))
	return manifest
}

func downloadLayer(image, digest, token, destDir string) {
	url := fmt.Sprintf("https://registry-1.docker.io/v2/library/%s/blobs/%s", image, digest)
	
	req, err := http.NewRequest("GET", url, nil)
	must(err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	must(err)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Failed to download layer %s\n", digest)
		return
	}

	filePath := filepath.Join(destDir, digest[7:20]+".tar.gz")
	out, err := os.Create(filePath)
	must(err)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	must(err)
	fmt.Printf("Saved to %s\n", filePath)
}