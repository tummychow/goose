// Package document provides the structures for Goose wiki pages (Document) and
// an abstraction over their storage (DocumentStore). It also provides various
// utility functions for the manipulation of Documents.
package document

import (
	"fmt"
	"time"
)

// MAX_CONTENT_SIZE is the maximum size of a Goose Document's Content in bytes.
// A Document's Content must be less than this length.
const MAX_CONTENT_SIZE = 512 * 1024

// DocumentStore represents Goose's Document storage backend.
//
// DocumentStore stores Documents in a Name-addressable store with immutable
// versions. Modifying a given Document never edits or overwrites the current
// version - instead, a new version is always created. Versions are identified
// and ordered by timestamp. For maintenance, a Document can be reverted to an
// older version (discarding newer ones), or its history can be truncated to a
// certain version to save space (discarding older ones).
//
// Several methods of the DocumentStore can return errors. Some errors are
// implementation-agnostic (eg DocumentNotFoundError) and all implementations
// must use these errors as noted. Any other error value has implementation-
// specific meaning and can appear at any time, in which case all other return
// values have implementation-specific meaning as well.
//
// DocumentStore requires eventual consistency across a single instance and
// across all copies thereof. Read-your-writes consistency is useful (required
// for tests) but not mandatory. Because Documents are immutable, there should
// be no conflicting writes.
//
// In the event of a timestamp collision, DocumentStore is expected to preserve
// both versions. If those versions are the most recent, then the behavior of
// Get is implementation-specific, but should be deterministic (ie a Get can
// return either version, but it should be the same version every time) and
// should match the behavior of GetAll (ie whichever version is returned by Get
// should also be the first version returned by GetAll).
//
// DocumentStores should be safe for concurrent access across goroutines.
// Instance-wide locking is an acceptable solution, since DocumentStores can be
// copied. Locking across copies should be avoided if possible.
//
// DocumentStore methods are expected to be synchronous, and block until the
// method behavior is completed. The meaning of "completed" depends on the
// implementation, however. For example, if Updating a Document to a replicated
// database cluster, you could return without waiting for any write ack, or
// waiting for a primary ack, or waiting for a majority ack. Depending on the
// implementation, any of these could be "complete".
//
// Because Document Timestamps require single-second precision, DocumentStore
// implementations must also support at least that much precision when storing
// and comparing timestamps.
//
// The Documents returned by a DocumentStore must be valid, satisfying all the
// guarantees defined for Document.
type DocumentStore interface {
	// Returns the Document specified by name, at its newest version.
	//
	// If the Document does not exist, the error return must be a non-nil
	// document.DocumentNotFoundError. In that case, the returned Document has
	// undefined value.
	Get(name string) (Document, error)

	// Returns all versions of the Document specified by name, in order from
	// newest (index 0) to oldest (index n-1).
	//
	// If the Document does not exist, the error return must be a non-nil
	// document.DocumentNotFoundError. In that case, the returned slice must be
	// empty.
	GetAll(name string) ([]Document, error)

	// Creates a new version of the Document specified by name, containing the
	// specified content. The int return is the version number of the created
	// Document. The Timestamp of the created version is determined by the
	// DocumentStore.
	//
	// Update can be invoked for Documents that do not exist, in which case the
	// first version is created. Update never modifies an old version of an
	// existing Document.
	//
	// A DocumentStore must support at least document.MAX_CONTENT_SIZE bytes of
	// content as an argument to this function. Passing a larger string must
	// return a non-nil DocumentTooLargeError.
	Update(name, content string) (int, error)

	// Reverts the Document specified by name, to the specified time. All
	// versions from that time or later will be discarded. Returns the number
	// of versions discarded.
	//
	// If the timestamp has high precision (finer than one second), the extra
	// precision may be ignored by the implementation. DocumentStore is only
	// required to support single-second precision; any finer than that is
	// not guaranteed.
	//
	// A Document can be completely deleted by Reverting it to the Timestamp of
	// its oldest version (or any earlier time).
	//
	// Reverting a Document that does not exist must have no effect. The
	// returns must be zero, and a non-nil document.DocumentNotFoundError.
	Revert(name string, timestamp time.Time) (int, error)

	// Truncates history for the Document specified by name, to the specified
	// time. All versions from that time or earlier will be discarded. Returns
	// the number of versions discarded.
	//
	// If the timestamp has high precision (finer than one second), the extra
	// precision may be ignored by the implementation. DocumentStore is only
	// required to support single-second precision; any finer than that is
	// not guaranteed.
	//
	// A Document can be completely deleted by Truncating it to the Timestamp
	// of its newest version (or any later time).
	//
	// Truncating a Document that does not exist must have no effect. The
	// returns must be zero, and a non-nil document.DocumentNotFoundError.
	Truncate(name string, timestamp time.Time) (int, error)

	// Returns a new DocumentStore instance that uses the same underlying
	// storage as the receiver. Copying a DocumentStore should be a lightweight
	// operation.
	//
	// The receiver and return value must have the same implementation type. In
	// addition, they should share the actual data and be eventually consistent
	// with each other. For example, if the receiver is using a connection to a
	// MongoDB replica set, its copies should use the same replica set, but
	// possibly with different TCP connections, or to different members of the
	// set.
	//
	// The interface does not impose a limit on the number of copies allowed.
	// Implementations may define their own limit, and impose it via an error
	// return when the limit is exceeded.
	Copy() (DocumentStore, error)

	// Closes the DocumentStore, allowing it to release any internal resources.
	//
	// After Closing a DocumentStore, any further DocumentStore method calls
	// must return a non-nil ClosedError.
	//
	// Closing a DocumentStore that is already Closed should have no effect.
	Close()
}

// Document represents a single version of a single Goose wiki page. Every page
// is UTF-8 Markdown. The storage and persistence of Documents is handled by a
// DocumentStore.
type Document struct {
	// Name is the key by which this Document can be retrieved from its
	// DocumentStore. A Name is unique across a DocumentStore instance and its
	// copies (ignoring versions).
	//
	// The format of Name is a nonempty sequence of segments. Each segment
	// consists of a slash, followed by at least one non-slash printable ASCII
	// character (ie any character in the range \x20-\x2E or \x30-\x7E).
	// Furthermore, a segment may not be the strings "/." or "/..".
	//
	// The Name is part of the URL used to access the Document. However, do not
	// %-encode the characters of the Name. In addition, the last slash-
	// separated segment in the string is the Document's title.
	Name string

	// Content is the body of the Document, in UTF-8 Markdown. See package
	// github.com/tummychow/goose/markdown for details on the expected syntax.
	//
	// The maximum size of a valid Content string is document.MAX_CONTENT_SIZE.
	// There are no other technical restrictions on this variable.
	Content string

	// Timestamp is the time at which this Document was added to its
	// DocumentStore. It must be a UTC timestamp with at least single-second
	// precision.
	//
	// The actual meaning of Timestamp is implementation-specific. It depends
	// on the DocumentStore from which this Document originated. In general, it
	// approximately reflects when Update was called to add this Document to
	// the DocumentStore.
	Timestamp time.Time
}

// DocumentNotFoundError is the error returned by a DocumentStore when an
// operation is attempted against a Document that does not exist.
type DocumentNotFoundError struct {
	// Name is the Name of the nonexistent Document that caused the error.
	Name string
}

func (e DocumentNotFoundError) Error() string {
	return fmt.Sprintf("goose/document: document %q not found", e.Name)
}

// ClosedError is the error returned when a method call is invoked on a
// DocumentStore that was already closed.
type ClosedError string

func (e ClosedError) Error() string {
	return string(e)
}

// DocumentTooLargeError is the error returned when a Document has too much
// content, exceeding document.MAX_CONTENT_SIZE.
type DocumentTooLargeError struct {
	Size int
}

func (e DocumentTooLargeError) Error() string {
	return fmt.Sprintf("goose/document: content is %v bytes (%v bytes too long)", e.Size, e.Size-MAX_CONTENT_SIZE)
}
