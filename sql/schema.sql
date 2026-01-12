IF OBJECT_ID ('Apartment_IDs', 'U') IS NULL BEGIN
CREATE TABLE Apartment_IDs (
    id INT PRIMARY KEY,
    is_scraped BIT DEFAULT 0,
    created_at DATETIME DEFAULT GETDATE ()
);

END;

IF OBJECT_ID ('Apartments', 'U') IS NULL BEGIN
CREATE TABLE Apartments (
    id INT PRIMARY KEY,
    source VARCHAR(50),
    city CHAR(6) DEFAULT 'Cracow',
    district NVARCHAR (50),
    latitude DECIMAL(15, 12),
    longitude DECIMAL(15, 12),
    total_price DECIMAL(15, 2),
    square_footage DECIMAL(10, 2),
    rooms INT,
    floor INT,
    total_floors INT,
    finishing_state NVARCHAR (50),
    market_type NVARCHAR (20),
    advertiser_type NVARCHAR (30),
    build_year INT,
    has_elevator BIT,
    heating_type NVARCHAR (50),
    avaiable_from DATE,
    created_at DATETIME DEFAULT GETDATE ()
);

END;