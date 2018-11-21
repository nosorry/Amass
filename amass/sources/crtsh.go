// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package sources

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/OWASP/Amass/amass/core"
	"github.com/OWASP/Amass/amass/utils"
)

type crtData struct {
	IssuerID          int    `json:"issuer_ca_id"`
	IssuerName        string `json:"issuer_name"`
	Name              string `json:"name_value"`
	MinCertID         int    `json:"min_cert_id"`
	MinEntryTimestamp string `json:"min_entry_timestamp"`
	NotBefore         string `json:"not_before"`
	NotAfter          string `json:"not_after"`
}

// Crtsh is the AmassService that handles access to the Crtsh data source.
type Crtsh struct {
	core.BaseAmassService

	SourceType string
}

// NewCrtsh returns he object initialized, but not yet started.
func NewCrtsh(e *core.Enumeration) *Crtsh {
	c := &Crtsh{SourceType: core.CERT}

	c.BaseAmassService = *core.NewBaseAmassService(e, "Crtsh", c)
	return c
}

// OnStart implements the AmassService interface
func (c *Crtsh) OnStart() error {
	c.BaseAmassService.OnStart()

	go c.startRootDomains()
	return nil
}

// OnStop implements the AmassService interface
func (c *Crtsh) OnStop() error {
	c.BaseAmassService.OnStop()
	return nil
}

func (c *Crtsh) startRootDomains() {
	// Look at each domain provided by the config
	for _, domain := range c.Enum().Config.Domains() {
		c.executeQuery(domain)
	}
}

func (c *Crtsh) executeQuery(domain string) {
	url := c.getURL(domain)
	page, err := utils.RequestWebPage(url, nil, nil, "", "")
	if err != nil {
		c.Enum().Log.Printf("%s: %s: %v", c.String(), url, err)
		return
	}

	c.SetActive()
	lines := json.NewDecoder(strings.NewReader(page))
	for {
		var line crtData
		if err := lines.Decode(&line); err == io.EOF {
			break
		} else if err != nil {
			c.Enum().Log.Printf("%s: %s: %v", c.String(), url, err)
			continue
		}

		req := &core.AmassRequest{
			Name:   cleanName(line.Name),
			Domain: domain,
			Tag:    c.SourceType,
			Source: c.String(),
		}

		if c.Enum().DupDataSourceName(req) {
			continue
		}
		c.Enum().Bus.Publish(core.NEWNAME, req)
	}
}

func (c *Crtsh) getURL(domain string) string {
	return "https://crt.sh/?q=%25." + domain + "&output=json"
}
