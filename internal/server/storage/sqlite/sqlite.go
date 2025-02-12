package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sync"

	"github.com/dvaxert/mdm/internal/domain/models"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
	mu sync.Mutex
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// В полноценном проекте я бы предпочел использовать миграции
	// однако в данном случае ограничимся созданием таблиц при создании storage
	err = initDb(db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) RegisterDevice(ctx context.Context, device_uuid uuid.UUID, device_type models.DeviceType) (int64, error) {
	const op = "storage.sqlite.Register"

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO devices(uuid, type) VALUES(?,?);")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, device_uuid.String(), device_type)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = tx.Prepare(
		`INSERT OR IGNORE INTO device_features(device_id, feature_id, state) 
		 VALUES(?,(SELECT id FROM features WHERE name = ?),?);`,
	)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	for k, v := range models.DefaultFeatures {
		_, err = stmt.ExecContext(ctx, id, k, v)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) Device(ctx context.Context, device_uuid uuid.UUID) (models.Device, error) {
	const op = "storage.sqlite.Device"

	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare("SELECT id, uuid, type FROM devices WHERE uuid = ?;")
	if err != nil {
		return models.Device{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, device_uuid)

	var device models.Device
	err = row.Scan(&device.Id, &device.Uuid, &device.Type)
	if err != nil {
		return models.Device{}, fmt.Errorf("%s: %w", op, err)
	}

	return device, nil
}

func (s *Storage) UpdateDeviceStatus(ctx context.Context, device_uuid uuid.UUID, location string, battery int) error {
	const op = "storage.sqlite.Register"

	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare(
		`INSERT OR REPLACE INTO device_statuses(device_id, location, battery) 
		 VALUES((SELECT id FROM devices WHERE uuid = ?),?,?);`,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, device_uuid, location, battery)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeviceStatus(ctx context.Context, device_uuid uuid.UUID) (models.DeviceStatus, error) {
	const op = "storage.sqlite.DeviceStatus"

	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare(
		`SELECT d.device_id, d.uuid, s.location, s.battery
		 FROM devices AS d
			JOIN device_statuses AS s
			ON d.id = s.device_id
		 WHERE d.uuid = ?`,
	)
	if err != nil {
		return models.DeviceStatus{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, device_uuid)

	var info models.DeviceStatus
	err = row.Scan(&info.DeviceId, &info.DeviceUuid, &info.Location, &info.Location)
	if err != nil {
		return models.DeviceStatus{}, fmt.Errorf("%s: %w", op, err)
	}

	return info, nil
}

func (s *Storage) UpdateDeviceFeature(ctx context.Context, device_uuid uuid.UUID, feature string, state bool) error {
	const op = "storage.sqlite.UpdateDeviceFeature"

	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare(
		`INSERT OR REPLACE INTO device_features(device_id, feature_id, state) 
		 VALUES(
		 	(SELECT id FROM devices WHERE uuid = ?), 
			(SELECT id FROM features WHERE name = ?), 
			?
		);`,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, device_uuid, feature, state)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeviceFeatures(ctx context.Context, device_uuid uuid.UUID) (models.DeviceFeatures, error) {
	const op = "storage.sqlite.DeviceFeatures"

	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare("SELECT id FROM devices WHERE uuid = ?;")
	if err != nil {
		return models.DeviceFeatures{}, fmt.Errorf("%s: %w", op, err)
	}

	var device_id int64
	row := stmt.QueryRowContext(ctx, device_uuid)
	row.Scan(&device_id)
	if err != nil {
		return models.DeviceFeatures{}, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = s.db.Prepare(
		`SELECT f.name, d.state 
		 FROM device_features AS d
			JOIN features AS f
			ON d.feature_id = f.id
		 WHERE d.device_id = ?;`,
	)
	if err != nil {
		return models.DeviceFeatures{}, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, device_id)
	if err != nil {
		return models.DeviceFeatures{}, fmt.Errorf("%s: %w", op, err)
	}

	features := models.DeviceFeatures{
		DeviceId: device_id,
		Features: make(map[string]bool),
	}

	for rows.Next() {
		var (
			name  string
			state bool
		)
		err = rows.Scan(&name, &state)
		if err != nil {
			return models.DeviceFeatures{}, fmt.Errorf("%s: %w", op, err)
		}

		features.Features[name] = state
	}

	return features, nil
}

func (s *Storage) DeviceList(ctx context.Context) ([]models.Device, error) {
	const op = "storage.sqlite.DeviceList"

	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare("SELECT COUNT(*) FROM devices;")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx)

	var count int
	if err = row.Scan(&count); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = s.db.Prepare("SELECT id, uuid, type FROM devices;")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make([]models.Device, 0, count)
	for rows.Next() {
		var d models.Device
		rows.Scan(&d.Id, &d.Uuid, &d.Type)

		result = append(result, d)
	}

	return result, nil
}

func (s *Storage) DeviceStatusList(ctx context.Context) ([]models.DeviceStatus, error) {
	const op = "storage.sqlite.DeviceStatusList"

	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare("SELECT COUNT(*) FROM devices;")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx)

	var count int
	if err = row.Scan(&count); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = s.db.Prepare(
		`SELECT s.device_id, d.uuid, s.location, s.battery
		 FROM device_statuses AS s
			JOIN devices AS d
			ON d.id == s.device_id;`,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make([]models.DeviceStatus, 0, count)
	for rows.Next() {
		var ds models.DeviceStatus
		rows.Scan(&ds.DeviceId, &ds.DeviceUuid, &ds.Location, &ds.Battery)

		result = append(result, ds)
	}

	return result, nil
}

func (s *Storage) DeviceFeaturesList(ctx context.Context) ([]models.DeviceFeatures, error) {
	const op = "storage.sqlite.DeviceFeaturesList"

	s.mu.Lock()
	defer s.mu.Unlock()

	stmt, err := s.db.Prepare("SELECT COUNT(*) FROM devices;")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx)

	var count int
	if err = row.Scan(&count); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = s.db.Prepare(
		`SELECT d.id, d.uuid, f.name, df.state 
		 FROM device_features AS df
			JOIN features AS f
			ON df.feature_id = f.id
			JOIN devices AS d
			ON df.device_id = d.id
		ORDER BY d.id;`,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make([]models.DeviceFeatures, 0, count)
	for rows.Next() {
		var (
			id            int64
			uuidStr       string
			feature_name  string
			feature_state bool
		)
		rows.Scan(&id, &uuidStr, &feature_name, &feature_state)

		i := slices.IndexFunc(result, func(df models.DeviceFeatures) bool { return df.DeviceId == id })
		if i == -1 {
			uuid, err := uuid.Parse(uuidStr)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			df := models.DeviceFeatures{
				DeviceId:   id,
				DeviceUuid: uuid,
				Features:   make(map[string]bool),
			}
			df.Features[feature_name] = feature_state

			result = append(result, df)
		} else {
			result[i].Features[feature_name] = feature_state
		}
	}

	return result, nil
}

func initDb(db *sql.DB) error {
	const op = "storage.sqlite.Init"

	_, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			uuid TEXT NOT NULL UNIQUE,
			type INTEGER NOT NULL
		);`,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS features (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);`,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`INSERT OR IGNORE INTO features(name) VALUES('camera'),('storage');`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS device_features (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id INTEGER NOT NULL,
			feature_id INTEGER NOT NULL,
			state INTEGER NOT NULL CHECK (state IN(0, 1)),
			CONSTRAINT device_features_devices_id_fk
				FOREIGN KEY(device_id)
				REFERENCES devices(id),
			CONSTRAINT device_features_features_id_fk
				FOREIGN KEY(feature_id)
				REFERENCES features(id),
			UNIQUE(device_id, feature_id)
		);`,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS device_statuses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id INTEGER NOT NULL UNIQUE,
			location TEXT NOT NULL,
			battery INT NOT NULL CHECK (battery BETWEEN 0 AND 100),
			CONSTRAINT device_statuses_devices_id_fk
				FOREIGN KEY(device_id)
				REFERENCES devices(id)
		);`,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
