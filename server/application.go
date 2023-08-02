package server

// Upload tar file and return the file name
// POST /application/deploy/upload

// Dockerconfig generate from git repo
// POST /application/deploy/dockerconfig/generate/git

// Dockerconfig generate from source code
// POST /application/deploy/dockerconfig/generate/tarball

// Deploy application
// POST /application/deploy
// Data :
// - git / tarball
// - dockerconfig
// - env variables
// - build args
// - image

// GET /applications
// GET /application/:id
// GET /application/:id/logs
// GET /application/:id/resources
// PUT /application/:id