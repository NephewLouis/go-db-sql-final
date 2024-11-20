package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

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
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	parAdd, err := store.Add(parcel)
	if err != nil {
		t.Fatal(err)
	}

	parGet, err := store.Get(parAdd)
	if err != nil {
		t.Fatal(err)
	}

	if parGet.Number != parcel.Number ||
		parGet.Client != parcel.Client ||
		parGet.Status != parcel.Status ||
		parGet.Address != parcel.Address {
		t.Errorf("Полученные данные не совпадают")
	}

	err = store.Delete(parAdd)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Get(parAdd)
	if err == nil {
		t.Fatalf("Посылка %d была удалена", parAdd)
	}
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	parAdd, err := store.Add(parcel)
	if err != nil {
		t.Fatal(err)
	}

	newAddress := "new test address"
	parcel.Number = parAdd
	parcel.Address = newAddress

	err = store.SetAddress(parcel.Number, parcel.Address)
	if err != nil {
		t.Fatal(err)
	}

	parCheck, err := store.Get(parAdd)
	if err != nil {
		t.Fatal(err)
	}

	if parCheck.Address != parcel.Address {
		t.Error("Адрес не обновился")
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	parAdd, err := store.Add(parcel)
	if err != nil {
		t.Fatal(err)
	}

	err = store.SetStatus(parcel.Number, parcel.Status)
	if err != nil {
		t.Fatal(err)
	}

	parCheck, err := store.Get(parAdd)
	if err != nil {
		t.Fatal(err)
	}

	if parCheck.Status != parcel.Status {
		t.Error("Статус не обновился")
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
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
	if err != nil {
		t.Fatal(err)
	}

	if len(parcels) != len(storedParcels) {
		t.Fatal("Количество полученных посылок не совпадает с количеством добавленных")
	}

	for _, parcel := range storedParcels {
		prod, exists := parcelMap[parcel.Client]
		if !exists {
			t.Errorf("Неизвестная посылка %d", parcel.Number)
			continue
		}

		if prod.Client != parcel.Client {
			t.Fatal("Найдено несоответствие клиента")
		}

		if prod.Status != parcel.Status {
			t.Fatal("Найдено несоответствие статуса")
		}

		if prod.Address != parcel.Address {
			t.Fatal("Найдено несоответствие статуса")
		}
	}
}
