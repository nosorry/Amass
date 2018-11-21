// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package sources

import (
	"fmt"

	"github.com/OWASP/Amass/amass/core"
	"github.com/OWASP/Amass/amass/utils"
)

// Exalead is the AmassService that handles access to the Exalead data source.
type PTRArchive struct {
	core.BaseAmassService

	SourceType string
}

// NewPTRArchive returns he object initialized, but not yet started.
func NewPTRArchive(e *core.Enumeration) *PTRArchive {
	p := &PTRArchive{SourceType: core.SCRAPE}

	p.BaseAmassService = *core.NewBaseAmassService(e, "PTRArchive", p)
	return p
}

// OnStart implements the AmassService interface
func (p *PTRArchive) OnStart() error {
	p.BaseAmassService.OnStart()

	go p.startRootDomains()
	return nil
}

// OnStop implements the AmassService interface
func (p *PTRArchive) OnStop() error {
	p.BaseAmassService.OnStop()
	return nil
}

func (p *PTRArchive) startRootDomains() {
	// Look at each domain provided by the config
	for _, domain := range p.Enum().Config.Domains() {
		p.executeQuery(domain)
	}
}

func (p *PTRArchive) executeQuery(domain string) {
	url := p.getURL(domain)
	page, err := utils.RequestWebPage(url, nil, nil, "", "")
	if err != nil {
		p.Enum().Log.Printf("%s: %s: %v", p.String(), url, err)
		return
	}

	p.SetActive()
	re := p.Enum().Config.DomainRegex(domain)
	for _, sd := range re.FindAllString(page, -1) {
		req := &core.AmassRequest{
			Name:   cleanName(sd),
			Domain: domain,
			Tag:    p.SourceType,
			Source: p.String(),
		}

		if p.Enum().DupDataSourceName(req) {
			continue
		}
		p.Enum().Bus.Publish(core.NEWNAME, req)
	}
}

func (p *PTRArchive) getURL(domain string) string {
	format := "http://ptrarchive.com/tools/search3.htm?label=%s&date=ALL"

	return fmt.Sprintf(format, domain)
}
