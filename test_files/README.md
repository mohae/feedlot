// This directory contains files for testing. It also is where tests write out
// files, when necessary. After a test run, prior to exiting test, the files 
// created during testing will be deleted. This may be moved to the system's
// temp directory, where it probably should be, but this is how it was written
// to start.
//
// The files contained within, with the exception of the TOML files in the conf
// directories, are not usable and contain the minimal needed for testing, if
// they contain enything. The tests also depend on these files so they should 
// not be modified unless necessary to fix the test or because of code changes.
