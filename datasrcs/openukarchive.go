// Copyright 2017 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package datasrcs

import (
	"context"
	"fmt"
	"time"

	"github.com/OWASP/Amass/v3/config"
	"github.com/OWASP/Amass/v3/eventbus"
	"github.com/OWASP/Amass/v3/requests"
	"github.com/OWASP/Amass/v3/systems"
)

// OpenUKArchive is the Service that handles access to the OpenUKArchive data source.
type OpenUKArchive struct {
	requests.BaseService

	SourceType string
	domain     string
	baseURL    string
}

// NewOpenUKArchive returns he object initialized, but not yet started.
func NewOpenUKArchive(sys systems.System) *OpenUKArchive {
	o := &OpenUKArchive{
		SourceType: requests.ARCHIVE,
		domain:     "webarchive.org.uk",
		baseURL:    "http://www.webarchive.org.uk/wayback/archive",
	}

	o.BaseService = *requests.NewBaseService(o, "OpenUKArchive")
	return o
}

// Type implements the Service interface.
func (o *OpenUKArchive) Type() string {
	return o.SourceType
}

// OnStart implements the Service interface.
func (o *OpenUKArchive) OnStart() error {
	o.BaseService.OnStart()

	o.SetRateLimit(time.Second)
	return nil
}

// OnDNSRequest implements the Service interface.
func (o *OpenUKArchive) OnDNSRequest(ctx context.Context, req *requests.DNSRequest) {
	cfg := ctx.Value(requests.ContextConfig).(*config.Config)
	bus := ctx.Value(requests.ContextEventBus).(*eventbus.EventBus)
	if cfg == nil || bus == nil {
		return
	}

	if req.Name == "" || req.Domain == "" {
		return
	}

	if !cfg.IsDomainInScope(req.Name) {
		return
	}

	o.CheckRateLimit()

	names, err := crawl(ctx, o.baseURL, o.domain, req.Name, req.Domain)
	if err != nil {
		bus.Publish(requests.LogTopic, eventbus.PriorityHigh, fmt.Sprintf("%s: %v", o.String(), err))
		return
	}

	for _, name := range names {
		bus.Publish(requests.NewNameTopic, eventbus.PriorityHigh, &requests.DNSRequest{
			Name:   cleanName(name),
			Domain: req.Domain,
			Tag:    o.SourceType,
			Source: o.String(),
		})
	}
}
