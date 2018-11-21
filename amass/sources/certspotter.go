// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package sources

import (
	"fmt"

	"github.com/OWASP/Amass/amass/core"
	"github.com/OWASP/Amass/amass/utils"
)

// CertSpotter is the AmassService that handles access to the CertSpotter data source.
type CertSpotter struct {
	core.BaseAmassService

	SourceType string
}

// NewCertSpotter returns he object initialized, but not yet started.
func NewCertSpotter(e *core.Enumeration) *CertSpotter {
	c := &CertSpotter{SourceType: core.CERT}

	c.BaseAmassService = *core.NewBaseAmassService(e, "CertSpotter", c)
	return c
}

// OnStart implements the AmassService interface
func (c *CertSpotter) OnStart() error {
	c.BaseAmassService.OnStart()

	go c.startRootDomains()
	return nil
}

// OnStop implements the AmassService interface
func (c *CertSpotter) OnStop() error {
	c.BaseAmassService.OnStop()
	return nil
}

func (c *CertSpotter) startRootDomains() {
	// Look at each domain provided by the config
	for _, domain := range c.Enum().Config.Domains() {
		c.executeQuery(domain)
	}
}

func (c *CertSpotter) executeQuery(domain string) {
	url := c.getURL(domain)
	page, err := utils.RequestWebPage(url, nil, nil, "", "")
	if err != nil {
		c.Enum().Log.Printf("%s: %s: %v", c.String(), url, err)
		return
	}

	c.SetActive()
	re := c.Enum().Config.DomainRegex(domain)
	for _, sd := range re.FindAllString(page, -1) {
		req := &core.AmassRequest{
			Name:   cleanName(sd),
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

func (c *CertSpotter) getURL(domain string) string {
	format := "https://certspotter.com/api/v0/certs?domain=%s"

	return fmt.Sprintf(format, domain)
}
