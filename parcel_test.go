package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var randSource = rand.NewSource(time.Now().UnixNano())
var randRange = rand.New(randSource)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {

	db, err := sql.Open("sqlite", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	p, err := store.Get(id)
	require.NoError(t, err)

	// Сравнение всей структуры (игнорируем Number и CreatedAt)
	expectedParcel := parcel
	expectedParcel.Number = id
	p.CreatedAt = expectedParcel.CreatedAt
	require.Equal(t, expectedParcel, p)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {

	db, err := sql.Open("sqlite", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)

	expectedParcel := parcel
	expectedParcel.Number = id
	expectedParcel.Address = newAddress
	p.CreatedAt = expectedParcel.CreatedAt
	require.Equal(t, expectedParcel, p)
}

func TestSetStatus(t *testing.T) {

	db, err := sql.Open("sqlite", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)

	expectedParcel := parcel
	expectedParcel.Number = id
	expectedParcel.Status = ParcelStatusSent
	p.CreatedAt = expectedParcel.CreatedAt
	require.Equal(t, expectedParcel, p)
}

func TestGetByClient(t *testing.T) {

	db, err := sql.Open("sqlite", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	client := randRange.Intn(10_000_000)
	parcelMap := make(map[int]Parcel)

	for i := 0; i < 3; i++ {
		p := getTestParcel()
		p.Client = client
		id, err := store.Add(p)
		require.NoError(t, err)
		p.Number = id
		parcelMap[id] = p
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, 3)

	for _, parcel := range storedParcels {
		expected, exists := parcelMap[parcel.Number]
		require.True(t, exists)

		parcel.CreatedAt = expected.CreatedAt
		require.Equal(t, expected, parcel)
	}
}
