package runners

import (
	"log"
	"sync"
	"time"

	"github.com/cha87de/kvmtop/config"
	"github.com/cha87de/kvmtop/connector"
	"github.com/cha87de/kvmtop/models"
	libvirt "github.com/libvirt/libvirt-go"
)

func initializeLookup(wg *sync.WaitGroup) {
	for n := -1; config.Options.Runs == -1 || n < config.Options.Runs; n++ {
		start := time.Now()
		lookup()
		nextRun := start.Add(time.Duration(config.Options.Frequency) * time.Second)
		time.Sleep(nextRun.Sub(time.Now()))
	}
	wg.Done()
}

func lookup() {
	// initialize models
	if models.Collection.Domains == nil {
		models.Collection.Domains = make(map[string]*models.Domain)
	}

	// query libvirt
	doms, err := connector.Libvirt.Connection.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		log.Printf("Cannot get list of domains form libvirt.")
		return
	}

	// create list of cached domains
	domIDs := make([]string, 0, len(models.Collection.Domains))
	for id := range models.Collection.Domains {
		domIDs = append(domIDs, id)
	}

	// update domain list
	for _, dom := range doms {
		domain, err := handleDomain(dom)
		if err != nil {
			continue
		}
		domIDs = removeFromArray(domIDs, domain.UUID)
	}

	// remove cached but not existent domains
	for _, id := range domIDs {
		delete(models.Collection.Domains, id)
	}

}

func handleDomain(dom libvirt.Domain) (*models.Domain, error) {
	uuid, err := dom.GetUUIDString()
	if err != nil {
		return nil, err
	}

	name, err := dom.GetName()
	if err != nil {
		return nil, err
	}

	if domain, ok := models.Collection.Domains[uuid]; ok {
		domain.Name = name
		models.Collection.Domains[uuid] = domain
	} else {
		models.Collection.Domains[uuid] = &models.Domain{
			UUID: string(uuid),
			Name: name,
		}
	}

	// call collector lookup functions
	domain := models.Collection.Domains[uuid]
	for _, collector := range models.Collection.Collectors {
		collector.Lookup(domain, dom)
	}

	return models.Collection.Domains[uuid], nil
}

func removeFromArray(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}