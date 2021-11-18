package jira

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pkg/errors"
)

var diffOnlyWhenCreate = func(k, old, new string, d *schema.ResourceData) bool {
	if old == "" {
		return old == new
	}
	return true
}

// resourceUser is used to define a JIRA issue
func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Delete: resourceUserDelete,
		Update: noOpsUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         false,
				DiffSuppressFunc: diffOnlyWhenCreate,
			},
			"display_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         false,
				DiffSuppressFunc: diffOnlyWhenCreate,
			},
		},
	}
}

func noOpsUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func getUserByKey(client *jira.Client, accountId string) (*jira.User, *jira.Response, error) {
	apiEndpoint := fmt.Sprintf("%s?accountId=%s", userAPIEndpoint, accountId)
	req, err := client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	user := new(jira.User)
	resp, err := client.Do(req, user)
	if err != nil {
		return nil, resp, jira.NewJiraError(resp, err)
	}
	return user, resp, nil
}

func deleteUserByKey(client *jira.Client, accountId string) (*jira.Response, error) {
	apiEndpoint := fmt.Sprintf("%s?accountId=%s", userAPIEndpoint, accountId)
	req, err := client.NewRequest("DELETE", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req, nil)
	if err != nil {
		return resp, jira.NewJiraError(resp, err)
	}
	return resp, nil
}

// resourceUserCreate creates a new jira user using the jira api
func resourceUserCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	user := new(jira.User)
	user.EmailAddress = d.Get("email").(string)
	user.DisplayName = d.Get("display_name").(string)
	createdUser, _, err := config.jiraClient.User.Create(user)
	if err != nil {
		return errors.Wrap(err, "Request failed")
	}

	d.SetId(createdUser.AccountID)

	return resourceUserRead(d, m)
}

// resourceUserRead reads issue details using jira api
func resourceUserRead(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	user, _, err := getUserByKey(config.jiraClient, d.Id())
	if err != nil {
		return errors.Wrap(err, "getting jira user failed")
	}

	d.Set("email", user.EmailAddress)
	d.Set("display_name", user.DisplayName)
	return nil
}

// resourceUserDelete deletes jira issue using the jira api
func resourceUserDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(*Config)

	_, err := deleteUserByKey(config.jiraClient, d.Id())

	if err != nil {
		return errors.Wrap(err, "Request failed")
	}

	return nil
}
