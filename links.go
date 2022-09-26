package main

import (
	"github.com/pkg/errors"
	"github.com/suapapa/site-ingress/ingress"
)

var (
	links     []*ingress.Link
	redirects = map[string]*ingress.Link{}
)

func updateLinks() error {
	if links != nil {
		return nil
	}

	ls, err := ingress.LoadLinksConf(linksConf)
	if err != nil {
		return errors.Wrap(err, "fail to get links")
	}
	// update redirect map
	for k := range redirects {
		delete(redirects, k)
	}
	for i, l := range ls {
		redirects[l.Name] = ls[i]
	}

	links = ls

	return nil
}
