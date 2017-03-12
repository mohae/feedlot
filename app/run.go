package app

import "github.com/mohae/feedlot/log"

// Run takes a list of build list names and generates all of the Packer
// templates associated with them.
func Run(listNames ...string) ([]string, []error) {
	log.Infof("run: build %d lists", len(listNames))
	// load the build lists
	bl := BuildLists{}
	err := bl.Load("")
	if err != nil {
		return nil, []error{err}
	}
	// make sure the lists all exist
	var errs []error
	var lists []List
	for _, name := range listNames {
		log.Debugf("%s: get list", name)
		l, err := bl.Get(name)
		if err != nil {
			errs = append(errs, err)
			log.Errorf("%s: get list: %s", err)
			continue
		}
		lists = append(lists, l)
		log.Debugf("%s: got list", name)
	}
	// if there were any errors on finding the build lists, don't do any processing.
	if errs != nil {
		log.Debug("run: exiting: errors occurred while retrieving build list info")
		return nil, errs
	}
	// Go through them and Build the builds in each list.
	// TODO: make this concurrent once concurrent generation of builds is stable
	messages := make([]string, len(listNames))
	for i, v := range lists {
		log.Debugf("run: build lists: %v", v.Builds)
		messages[i], err = BuildBuilds(v.Builds...)
		if err != nil {
			log.Infof("run: %s", err)
			errs = append(errs, err)
		} else {
			log.Infof("run: %s", messages[i])
		}
	}
	return messages, errs
}
