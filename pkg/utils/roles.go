package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/lenninlasd/Engine/types"
)

func GetUserRole(userId string) (role string, err error) {
	var tokenString = os.Getenv("SUPA_SERVICE_ROLE_TOKEN")
	var supabaseProjectId = os.Getenv("SUPA_PROJECT_ID")
	supabaseApiURL := fmt.Sprintf("https://%s.supabase.co/rest/v1/users?select=role&id=eq.%s", supabaseProjectId, userId)

	body, err := getApi(supabaseApiURL, tokenString)
	if err != nil {
		return "", err
	}

	var userRoles []types.UserRole
	err = json.Unmarshal(body, &userRoles)
	if err != nil {
		return "", err
	}

	if len(userRoles) == 0 {
		return "", nil
	}

	return userRoles[0].Role, nil
}

func getApi(url string, baseKey string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", baseKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", baseKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
