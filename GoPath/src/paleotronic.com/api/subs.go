package s8webclient

func (c *Client) SubscribeChannel(name string) {

	c.c.SendMessage(
		"SUB",
		[]byte(name),
		true,
	)

}

func (c *Client) UnsubscribeChannel(name string) {

	c.c.SendMessage(
		"USB",
		[]byte(name),
		true,
	)

}
