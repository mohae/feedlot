package app

// Run takes a list of build list names and generates all of the Packer
// templates associated with them.
func Run(listNames ...string) ([]string, []error) {
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
		l, err := bl.Get(name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		lists = append(lists, l)

	}
	// if there were any errors on finding the build lists, don't do any processing.
	if errs != nil {
		return nil, errs
	}
	// Go through them and Build the builds in each list.
	// TODO: make this concurrent once concurrent generation of builds is stable
	messages := make([]string, len(listNames))
	for i, v := range lists {
		messages[i], err = BuildBuilds(v.Builds...)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return messages, errs
}
