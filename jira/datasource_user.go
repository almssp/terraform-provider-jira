package jira

import (
	"fmt"
	"log"
    "time"
	jira "github.com/andygrunwald/go-jira"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
)

// resourceUser is used to define a JIRA issue
func datasourceUser() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceUserRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:             schema.TypeString,
				Required:         true,
			},
		},
	}
}

func getUserByQuery(client *jira.Client, email string) (*[]jira.User, *jira.Response, error) {
	apiEndpoint := fmt.Sprintf("%s/search?query=%s&maxResults=1", userAPIEndpoint, email)

	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	users := new([]jira.User)
    //Retries if empty
    maxtries := 2
    //First try
	resp, err := client.Do(req, users)
    for i := 1 ; i <= maxtries ; i++ {

	    if err != nil || len(*users) > 0 {
            log.Printf("Iteration Try: %d - Found match for: %s",i, email)
            break
        }

        log.Printf("Iteration Try: %d For: %s - Sleeping for 4 seconds",i, email)
        time.Sleep(4 * time.Second)
	    resp, err = client.Do(req, users)
    }
    if len(*users) < 1 {
        log.Printf("No user found for %s", email)

    }
	if err != nil {
		return nil, resp, jira.NewJiraError(resp, err)
	}
    
	return users, resp, nil
}

// datasourceUserRead reads search results
func datasourceUserRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)
    email := d.Get("email").(string)
    
	users, _, err := getUserByQuery(config.jiraClient, email)
	if err != nil {
		return errors.Wrap(err, "getting jira user failed")
	}
    
    for _, user := range *users {
        log.Printf("User found - %s: %s\n", user.AccountID, user.DisplayName)
	    d.Set("display_name", user.DisplayName)
        d.SetId(user.AccountID)
	}
   
	return nil
}
