package main

import (
	"database/sql"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	parAdd, err := store.Add(parcel)
	require.NoError(t, err)

	parGet, err := store.Get(parAdd)
	require.NoError(t, err)

	parcel.Number = parAdd
	assert.Equal(t, parGet, parcel, "Полученные данные не совпадают")

	err = store.Delete(parAdd)
	require.NoError(t, err)

	_, err = store.Get(parAdd)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows, "Посылка не найдена")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	parAdd, err := store.Add(parcel)
	require.NoError(t, err)

	log.Print(parcel.Address)

	newAddress := "new test address"

	err = store.SetAddress(parAdd, newAddress)
	require.NoError(t, err)

	parcel.Number = parAdd
	parcel.Address = newAddress

	parCheck, err := store.Get(parAdd)
	require.NoError(t, err)

	log.Print(parcel.Address)

	assert.Equal(t, parCheck.Address, parcel.Address, "Данные не обновлены")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	parAdd, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(parcel.Number, parcel.Status)
	require.NoError(t, err)

	parCheck, err := store.Get(parAdd)
	require.NoError(t, err)

	assert.Equal(t, parcel.Status, parCheck.Status, "Статус не обновился")

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Fatal(err)
		}
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	assert.Len(t, parcelMap, len(storedParcels),
		"Количество полученных посылок не совпадает с количеством добавленных")

	assert.ElementsMatch(t, parcels, storedParcels, "Списки посылок не совпадают")
}
