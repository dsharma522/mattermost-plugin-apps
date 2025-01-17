package appclient

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Client struct {
	*model.Client4
	*ClientPP
	userID string
}

func as(token string, cc apps.Context) *Client {
	return NewClient(token, cc.MattermostSiteURL)
}

func AsBot(cc apps.Context) *Client {
	client := as(cc.BotAccessToken, cc)
	client.userID = cc.BotUserID
	return client
}

func AsActingUser(cc apps.Context) *Client {
	client := as(cc.ActingUserAccessToken, cc)
	if cc.ActingUser != nil {
		client.userID = cc.ActingUser.Id
	}
	return client
}

func NewClient(token, mattermostSiteURL string) *Client {
	c := Client{
		ClientPP: NewAppsPluginAPIClient(mattermostSiteURL),
		Client4:  model.NewAPIv4Client(mattermostSiteURL),
	}
	c.Client4.SetOAuthToken(token)
	c.ClientPP.SetOAuthToken(token)
	return &c
}

func (c *Client) KVSet(prefix, id string, in interface{}) (bool, error) {
	changed, res, err := c.ClientPP.KVSet(prefix, id, in)
	if err != nil {
		return false, err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return false, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return changed, nil
}

func (c *Client) KVGet(prefix, id string, ref interface{}) error {
	res, err := c.ClientPP.KVGet(prefix, id, ref)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) KVDelete(prefix, id string) error {
	res, err := c.ClientPP.KVDelete(prefix, id)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) Subscribe(sub *apps.Subscription) error {
	res, err := c.ClientPP.Subscribe(sub)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) GetSubscriptions() ([]apps.Subscription, error) {
	subs, res, err := c.ClientPP.GetSubscriptions()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return subs, nil
}

func (c *Client) Unsubscribe(sub *apps.Subscription) error {
	res, err := c.ClientPP.Unsubscribe(sub)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) StoreOAuth2App(oauth2App apps.OAuth2App) error {
	res, err := c.ClientPP.StoreOAuth2App(oauth2App)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) StoreOAuth2User(ref interface{}) error {
	res, err := c.ClientPP.StoreOAuth2User(ref)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) GetOAuth2User(ref interface{}) error {
	res, err := c.ClientPP.GetOAuth2User(ref)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return errors.Errorf("returned with status %d", res.StatusCode)
	}

	return nil
}

func (c *Client) Call(creq apps.CallRequest) (*apps.CallResponse, error) {
	cresp, res, err := c.ClientPP.Call(creq)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return cresp, nil
}

func (c *Client) CreatePost(post *model.Post) (*model.Post, error) {
	createdPost, res, err := c.Client4.CreatePost(post)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("returned with status %d", res.StatusCode)
	}

	return createdPost, nil
}

func (c *Client) DM(userID string, format string, args ...interface{}) (*model.Post, error) {
	return c.DMPost(userID, &model.Post{
		Message: fmt.Sprintf(format, args...),
	})
}

func (c *Client) DMPost(userID string, post *model.Post) (*model.Post, error) {
	if c.userID == "" {
		return nil, errors.New("empty sender user_id, perhaps Call does not expand acting_user")
	}

	channel, res, err := c.CreateDirectChannel(c.userID, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get direct channel")
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("returned with status %d", res.StatusCode)
	}

	post.ChannelId = channel.Id
	return c.CreatePost(post)
}
