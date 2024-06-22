package repo

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"
	"strings"
)

const (
	dirObjects = ".got/objects"
	dirRefs    = ".got/refs"
	dirHeads   = ".got/refs/heads"
	dirGot     = ".got"
)

func Init(root string) error {
	dir := path.Join(root, dirGot)
	if _, err := os.Stat(dir); err == nil {
		log.Printf("got repository %v already exists", dir)
		return nil
	}

	dirs := []string{dirGot, dirObjects, dirRefs, dirHeads}
	for _, d := range dirs {
		err := os.MkdirAll(path.Join(root, d), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %v, %w", d, err)
		}
	}

	return nil
}

func RootDir(cwd string) (string, error) {
	dir := path.Join(cwd, dirGot)
	_, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		strings.Split(dir, "")
		if cwd == "/" {
			return "", os.ErrNotExist
		}

		return RootDir(path.Dir(cwd))
	} else if err != nil {
		return "", err
	} else {
		return cwd, nil
	}
}

type ObjectType string

const (
	ObjTypBlob   = "blob"
	ObjTypCommit = "commit"
	ObjTypTag    = "tag"
	ObjTypTree   = "tree"
)

type Object struct {
	Type ObjectType
	Data []byte
	Hash string
}

func NewObject(typ ObjectType, r io.Reader) (*Object, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	prefix := []byte(fmt.Sprintf("%v\x00", typ))
	bytes = append(prefix, bytes...)

	hash := sha1.New()
	_, err = hash.Write(bytes)
	if err != nil {
		return nil, err
	}

	return &Object{
		Type: typ,
		Data: bytes,
		Hash: hex.EncodeToString(hash.Sum(nil)),
	}, nil
}

func NewRepository(cwd string) (*Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory, %w", err)
	}

	dir, err := RootDir(cwd)
	if err != nil {
		return nil, fmt.Errorf("could not get root repository: %w", err)
	}
	return &Repository{
		fs: os.DirFS(dir),
	}, nil
}

type Repository struct {
	fs fs.FS
}

func (r *Repository) WriteObject(obj *Object) error {
	objPath := r.objectPath(obj.Hash)
	err := os.MkdirAll(path.Dir(objPath), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(objPath, obj.Data, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) ReadObject(hash string) (*Object, error) {
	objPath := r.objectPath(hash)
	b, err := os.ReadFile(objPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("not a valid object %v", hash)
	} else if err != nil {
		return nil, fmt.Errorf("could not read object: %w", err)
	}

	idx := slices.Index(b, 0x00)

	return &Object{
		Hash: hash,
		Data: b[idx+1:],
		Type: ObjectType(b[:idx]),
	}, nil
}

func (r *Repository) objectPath(hash string) string {
	return path.Join(dirObjects, hash[0:2], hash)
}

func (r *Repository) WriteTree(dir string) (*Object, error) {
	entries, err := fs.ReadDir(r.fs, dir)
	if err != nil {
		return nil, fmt.Errorf("can not read dir %v: %w", dir, err)
	}

	var tree strings.Builder
	for _, e := range entries {
		p := path.Join(dir, e.Name())
		if r.ignoredFile(p) {
			continue
		}

		var obj *Object
		var err error
		if e.IsDir() {
			obj, err = r.WriteTree(p)
		} else {
			f, err := r.fs.Open(p)
			if err != nil {
				return nil, fmt.Errorf("could not open file %v: %w", p, err)
			}
			obj, err = NewObject(ObjTypBlob, f)
			err = r.WriteObject(obj)
			if err != nil {
				return nil, fmt.Errorf("failed to save object ID %v: %w", obj.Hash, err)
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to has object %v: %w", p, err)
		}

		_, _ = tree.WriteString(fmt.Sprintf("%v %v %v\n", obj.Type, obj.Hash, e.Name()))
	}

	data := tree.String()
	obj, err := NewObject(ObjTypTree, strings.NewReader(data))
	err = r.WriteObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to save object ID %v: %w", obj.Hash, err)
	}
	return obj, nil
}

func (r *Repository) ignoredFile(p string) bool {
	return p == ".got" || p == ".git"
}
