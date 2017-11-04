// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

type ObjectType string

const (
	OBJECT_COMMIT ObjectType = "commit"
	OBJECT_TREE   ObjectType = "tree"
	OBJECT_BLOB   ObjectType = "blob"
	OBJECT_TAG    ObjectType = "tag"
)

// Returns true if object with given id exists in the repository.
func (repo *Repository) IsObjectExist(objectId string) bool {
	log("searching for object %s", objectId)
	_, err := NewCommand("cat-file", "-e", objectId).RunInDir(repo.Path)
	if err != nil {
		log("search error: %v", err)
	}
	return err == nil
}
