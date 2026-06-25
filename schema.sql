CREATE TABLE creneaux (
    id SERIAL PRIMARY KEY,
    heure VARCHAR(5) NOT NULL,
    disponible BOOLEAN DEFAULT true,
    client VARCHAR(100),
    date DATE NOT NULL DEFAULT CURRENT_DATE
);
