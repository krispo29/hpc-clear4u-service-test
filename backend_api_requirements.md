# Backend API Requirements - MAWB System Integration

## Overview

ต้องการพัฒนา API endpoints สำหรับเชื่อมต่อ MAWB Info กับ Cargo Manifest และ Draft MAWB โดยใช้ foreign key relationship

## Database Schema Requirements

### 1. MAWB Info Table (มีอยู่แล้ว)

```sql
CREATE TABLE tbl_mawb_info (
    uuid VARCHAR(36) PRIMARY KEY,
    mawb VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    service_type VARCHAR(100) NOT NULL,
    shipping_type VARCHAR(100) NOT NULL,
    chargeable_weight DECIMAL(10,2) DEFAULT 0,
    created_by VARCHAR(100) DEFAULT 'admin',
    status ENUM('Draft', 'Pending', 'Confirmed', 'Rejected') DEFAULT 'Draft',
    cargo_manifest_status ENUM('Draft', 'Confirmed', 'Rejected') DEFAULT 'Draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 2. Cargo Manifest Table (มีอยู่แล้ว)

```sql
CREATE TABLE cargo_manifest (
    uuid VARCHAR(36) PRIMARY KEY,
    mawb_info_uuid VARCHAR(36) NOT NULL,
    mawb_number VARCHAR(255) NOT NULL,
    port_of_discharge VARCHAR(255),
    flight_no VARCHAR(100),
    freight_date VARCHAR(50),
    shipper TEXT,
    consignee TEXT,
    total_ctn VARCHAR(100),
    transshipment TEXT,
    status ENUM('Draft', 'Confirmed', 'Rejected') DEFAULT 'Draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (mawb_info_uuid) REFERENCES mawb_info(uuid) ON DELETE CASCADE
);
```

### 3. Cargo Manifest Items Table (มีอยู่แล้ว)

```sql
CREATE TABLE cargo_manifest_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    cargo_manifest_uuid VARCHAR(36) NOT NULL,
    hawb_no VARCHAR(255),
    pkgs VARCHAR(100),
    gross_weight VARCHAR(100),
    destination VARCHAR(100),
    commodity TEXT,
    shipper_name_address TEXT,
    consignee_name_address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cargo_manifest_uuid) REFERENCES cargo_manifest(uuid) ON DELETE CASCADE
);
```

### 4. Draft MAWB Table (มีอยู่แล้ว)

```sql
CREATE TABLE draft_mawb (
    uuid VARCHAR(36) PRIMARY KEY,
    mawb_info_uuid VARCHAR(36) NOT NULL,
    customer_uuid VARCHAR(36),
    airline_logo VARCHAR(500),
    airline_name VARCHAR(255),
    mawb VARCHAR(255),
    hawb VARCHAR(255),
    shipper_name_and_address TEXT,
    awb_issued_by VARCHAR(255),
    consignee_name_and_address TEXT,
    issuing_carrier_agent_name VARCHAR(255),
    accounting_infomation TEXT,
    agents_iata_code VARCHAR(50),
    account_no VARCHAR(100),
    airport_of_departure VARCHAR(100),
    reference_number VARCHAR(100),
    optional_shipping_info1 VARCHAR(255),
    optional_shipping_info2 VARCHAR(255),
    routing_to VARCHAR(100),
    routing_by VARCHAR(100),
    destination_to1 VARCHAR(100),
    destination_by1 VARCHAR(100),
    destination_to2 VARCHAR(100),
    destination_by2 VARCHAR(100),
    currency VARCHAR(10),
    chgs_code VARCHAR(10),
    wt_val_ppd VARCHAR(10),
    wt_val_coll VARCHAR(10),
    other_ppd VARCHAR(10),
    other_coll VARCHAR(10),
    declared_val_carriage VARCHAR(100),
    declared_val_customs VARCHAR(100),
    airport_of_destination VARCHAR(100),
    requested_flight_date1 VARCHAR(100),
    requested_flight_date2 VARCHAR(100),
    amount_of_insurance VARCHAR(100),
    handling_infomation TEXT,
    sci VARCHAR(255),
    prepaid DECIMAL(10,2) DEFAULT 0,
    valuation_charge DECIMAL(10,2) DEFAULT 0,
    tax DECIMAL(10,2) DEFAULT 0,
    total_other_charges_due_agent DECIMAL(10,2) DEFAULT 0,
    total_other_charges_due_carrier DECIMAL(10,2) DEFAULT 0,
    total_prepaid DECIMAL(10,2) DEFAULT 0,
    currency_conversion_rates VARCHAR(255),
    signature1 VARCHAR(255),
    signature2_date DATE,
    signature2_place VARCHAR(255),
    signature2_issuing VARCHAR(255),
    shipping_mark TEXT,
    status ENUM('Draft', 'Confirmed', 'Rejected') DEFAULT 'Draft',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (mawb_info_uuid) REFERENCES mawb_info(uuid) ON DELETE CASCADE
);
```

### 5. Draft MAWB Items Table (มีอยู่แล้ว)

```sql
CREATE TABLE draft_mawb_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    draft_mawb_uuid VARCHAR(36) NOT NULL,
    pieces_rcp VARCHAR(100),
    gross_weight VARCHAR(100),
    kg_lb ENUM('kg', 'lb') DEFAULT 'kg',
    rate_class VARCHAR(255),
    total_volume VARCHAR(100),
    chargeable_weight VARCHAR(100),
    rate_charge DECIMAL(10,2) DEFAULT 0,
    total DECIMAL(10,2) DEFAULT 0,
    nature_and_quantity TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_mawb_uuid) REFERENCES draft_mawb(uuid) ON DELETE CASCADE
);
```

### 6. Draft MAWB Item Dimensions Table (มีอยู่แล้ว)

```sql
CREATE TABLE draft_mawb_item_dims (
    id INT AUTO_INCREMENT PRIMARY KEY,
    draft_mawb_item_id INT NOT NULL,
    length VARCHAR(50),
    width VARCHAR(50),
    height VARCHAR(50),
    count VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_mawb_item_id) REFERENCES draft_mawb_items(id) ON DELETE CASCADE
);
```

### 7. Draft MAWB Charges Table (มีอยู่แล้ว)

```sql
CREATE TABLE draft_mawb_charges (
    id INT AUTO_INCREMENT PRIMARY KEY,
    draft_mawb_uuid VARCHAR(36) NOT NULL,
    charge_key VARCHAR(255),
    charge_value DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (draft_mawb_uuid) REFERENCES draft_mawb(uuid) ON DELETE CASCADE
);
```

## API Endpoints Requirements

### 1. Cargo Manifest APIs

#### GET /v1/mawbinfo/{uuid}/cargo-manifest

**Purpose:** ดึงข้อมูล Cargo Manifest ที่เชื่อมกับ MAWB Info

**Request:**

```
GET /v1/mawbinfo/550e8400-e29b-41d4-a716-446655440000/cargo-manifest
Authorization: Bearer {token}
```

**Response Success (200):**

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "uuid": "cargo-manifest-uuid",
    "mawb_info_uuid": "550e8400-e29b-41d4-a716-446655440000",
    "mawbNumber": "020-35310671",
    "portOfDischarge": "MUNICH",
    "flightNo": "LH773",
    "freightDate": "16/05/2025",
    "shipper": "GUANGZHOU RICH SHIPPING...",
    "consignee": "SPEDITION F.R.E.I.T.A.N. GMBH...",
    "totalCtn": "127 CTN",
    "transshipment": "TRANSSHIPMENT CARGO...",
    "status": "Draft",
    "items": [
      {
        "hawbNo": "CFL2505006",
        "pkgs": "127",
        "grossWeight": "2,927 KG",
        "dst": "MUC",
        "commodity": "DRESS (HS CODE: 620443)...",
        "shipperNameAndAddress": "GUANGZHOU SEAFLOWER...",
        "consigneeNameAndAddress": "MARJO LEDER & TRACHT..."
      }
    ],
    "createdAt": "2025-01-08T10:30:00Z",
    "updatedAt": "2025-01-08T10:30:00Z"
  }
}
```

**Response Not Found (404):**

```json
{
  "code": 404,
  "message": "Cargo Manifest not found for this MAWB"
}
```

#### POST /v1/mawbinfo/{uuid}/cargo-manifest

**Purpose:** สร้างหรืออัปเดต Cargo Manifest

**Request:**

```json
{
  "mawbNumber": "020-35310671",
  "portOfDischarge": "MUNICH",
  "flightNo": "LH773",
  "freightDate": "16/05/2025",
  "shipper": "GUANGZHOU RICH SHIPPING INTL CO,LTD...",
  "consignee": "SPEDITION F.R.E.I.T.A.N. GMBH...",
  "totalCtn": "127 CTN",
  "transshipment": "TRANSSHIPMENT CARGO FROM SZX TO MUC...",
  "items": [
    {
      "hawbNo": "CFL2505006",
      "pkgs": "127",
      "grossWeight": "2,927 KG",
      "dst": "MUC",
      "commodity": "DRESS (HS CODE: 620443)...",
      "shipperNameAndAddress": "GUANGZHOU SEAFLOWER...",
      "consigneeNameAndAddress": "MARJO LEDER & TRACHT..."
    }
  ]
}
```

**Response Success (200/201):**

```json
{
  "code": 200,
  "message": "Cargo Manifest created/updated successfully",
  "data": {
    "uuid": "cargo-manifest-uuid",
    "mawb_info_uuid": "550e8400-e29b-41d4-a716-446655440000"
    // ... same structure as GET response
  }
}
```

### 2. Draft MAWB APIs

#### GET /v1/mawbinfo/{uuid}/draft-mawb

**Purpose:** ดึงข้อมูล Draft MAWB ที่เชื่อมกับ MAWB Info

**Request:**

```
GET /v1/mawbinfo/550e8400-e29b-41d4-a716-446655440000/draft-mawb
Authorization: Bearer {token}
```

**Response Success (200):**

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "uuid": "draft-mawb-uuid",
    "mawb_info_uuid": "550e8400-e29b-41d4-a716-446655440000",
    "customerUUID": "customer-uuid",
    "airlineLogo": "https://example.com/logo.png",
    "airlineName": "Thai Airways",
    "mawb": "020-35310671",
    "hawb": "CFL2505006",
    "shipperNameAndAddress": "GUANGZHOU RICH SHIPPING...",
    "consigneeNameAndAddress": "SPEDITION F.R.E.I.T.A.N. GMBH...",
    "items": [
      {
        "id": 1,
        "piecesRCP": "127",
        "grossWeight": "2927",
        "kgLb": "kg",
        "rateClass": "N",
        "totalVolume": "15.5",
        "chargeableWeight": "2927.00",
        "rateCharge": 2.5,
        "total": 7317.5,
        "natureAndQuantity": "DRESS, SKIRT, APRON",
        "dims": [
          {
            "length": "120",
            "width": "80",
            "height": "100",
            "count": "50"
          }
        ]
      }
    ],
    "charges": [
      {
        "key": "AWC",
        "value": 25.0
      }
    ],
    "status": "Draft",
    "createdAt": "2025-01-08T10:30:00Z",
    "updatedAt": "2025-01-08T10:30:00Z"
  }
}
```

#### POST /v1/mawbinfo/{uuid}/draft-mawb

**Purpose:** สร้างหรืออัปเดต Draft MAWB

**Request:**

```json
{
  "customerUUID": "customer-uuid",
  "airlineLogo": "https://example.com/logo.png",
  "airlineName": "Thai Airways",
  "mawb": "020-35310671",
  "hawb": "CFL2505006",
  "shipperNameAndAddress": "GUANGZHOU RICH SHIPPING...",
  "consigneeNameAndAddress": "SPEDITION F.R.E.I.T.A.N. GMBH...",
  "items": [
    {
      "piecesRCP": "127",
      "grossWeight": "2927",
      "kgLb": "kg",
      "rateClass": "N",
      "rateCharge": 2.5,
      "natureAndQuantity": "DRESS, SKIRT, APRON",
      "dims": [
        {
          "length": "120",
          "width": "80",
          "height": "100",
          "count": "50"
        }
      ]
    }
  ],
  "charges": [
    {
      "key": "AWC",
      "value": 25.0
    }
  ]
}
```

**Response Success (200/201):**

```json
{
  "code": 200,
  "message": "Draft MAWB created/updated successfully",
  "data": {
    "uuid": "draft-mawb-uuid"
  }
}
```

### 3. Status Management APIs

#### POST /v1/mawbinfo/{uuid}/cargo-manifest/confirm

**Purpose:** Confirm Cargo Manifest

**Response:**

```json
{
  "code": 200,
  "message": "Cargo Manifest confirmed successfully"
}
```

#### POST /v1/mawbinfo/{uuid}/cargo-manifest/reject

**Purpose:** Reject Cargo Manifest

#### GET /v1/mawbinfo/{uuid}/cargo-manifest/print

**Purpose:** Generate PDF for Cargo Manifest

**Response:** PDF file (application/pdf)

## Golang Implementation Guidelines

### 1. Struct Definitions

```go
type CargoManifest struct {
    UUID            string                `json:"uuid" db:"uuid"`
    MAWBInfoUUID    string                `json:"mawb_info_uuid" db:"mawb_info_uuid"`
    MAWBNumber      string                `json:"mawbNumber" db:"mawb_number"`
    PortOfDischarge string                `json:"portOfDischarge" db:"port_of_discharge"`
    FlightNo        string                `json:"flightNo" db:"flight_no"`
    FreightDate     string                `json:"freightDate" db:"freight_date"`
    Shipper         string                `json:"shipper" db:"shipper"`
    Consignee       string                `json:"consignee" db:"consignee"`
    TotalCtn        string                `json:"totalCtn" db:"total_ctn"`
    Transshipment   string                `json:"transshipment" db:"transshipment"`
    Status          string                `json:"status" db:"status"`
    Items           []CargoManifestItem   `json:"items"`
    CreatedAt       time.Time             `json:"createdAt" db:"created_at"`
    UpdatedAt       time.Time             `json:"updatedAt" db:"updated_at"`
}

type CargoManifestItem struct {
    ID                      int    `json:"id" db:"id"`
    CargoManifestUUID       string `json:"cargo_manifest_uuid" db:"cargo_manifest_uuid"`
    HAWBNo                  string `json:"hawbNo" db:"hawb_no"`
    Pkgs                    string `json:"pkgs" db:"pkgs"`
    GrossWeight             string `json:"grossWeight" db:"gross_weight"`
    Destination             string `json:"dst" db:"destination"`
    Commodity               string `json:"commodity" db:"commodity"`
    ShipperNameAndAddress   string `json:"shipperNameAndAddress" db:"shipper_name_address"`
    ConsigneeNameAndAddress string `json:"consigneeNameAndAddress" db:"consignee_name_address"`
}
```

### 2. Handler Functions

```go
// GET /v1/mawbinfo/{uuid}/cargo-manifest
func GetCargoManifest(c *gin.Context) {
    mawbUUID := c.Param("uuid")

    // Validate MAWB exists
    // Query cargo_manifest with items
    // Return response
}

// POST /v1/mawbinfo/{uuid}/cargo-manifest
func CreateOrUpdateCargoManifest(c *gin.Context) {
    mawbUUID := c.Param("uuid")
    var req CargoManifest

    // Bind JSON request
    // Validate MAWB exists
    // Begin transaction
    // Insert/Update cargo_manifest
    // Insert/Update cargo_manifest_items
    // Commit transaction
    // Return response
}
```

### 3. Database Operations

```go
func (r *CargoManifestRepository) GetByMAWBUUID(mawbUUID string) (*CargoManifest, error) {
    // Query cargo_manifest
    // Query cargo_manifest_items
    // Join and return
}

func (r *CargoManifestRepository) CreateOrUpdate(manifest *CargoManifest) error {
    // Begin transaction
    // Upsert cargo_manifest
    // Delete existing items
    // Insert new items
    // Commit transaction
}
```

## Validation Requirements

1. **MAWB UUID Validation:** ตรวจสอบว่า MAWB Info UUID มีอยู่จริง
2. **Required Fields:** mawbNumber ต้องไม่เป็นค่าว่าง
3. **Status Validation:** status ต้องเป็น enum ที่กำหนด
4. **Items Validation:** ตรวจสอบ items array structure

## Error Handling

1. **404:** MAWB Info ไม่พบ
2. **400:** Request body ไม่ถูกต้อง
3. **500:** Database error
4. **401/403:** Authentication/Authorization error

## Testing Requirements

1. Unit tests สำหรับ handlers
2. Integration tests สำหรับ database operations
3. API tests สำหรับ endpoints
4. Test cases สำหรับ error scenarios

## Performance Considerations

1. ใช้ database indexes สำหรับ foreign keys
2. Implement pagination สำหรับ large datasets
3. Consider caching สำหรับ frequently accessed data
4. Optimize queries with proper JOINs

## Draft MAWB Struct Definitions

```go
type DraftMAWB struct {
    UUID                        string           `json:"uuid" db:"uuid"`
    MAWBInfoUUID               string           `json:"mawb_info_uuid" db:"mawb_info_uuid"`
    CustomerUUID               string           `json:"customerUUID" db:"customer_uuid"`
    AirlineLogo                string           `json:"airlineLogo" db:"airline_logo"`
    AirlineName                string           `json:"airlineName" db:"airline_name"`
    MAWB                       string           `json:"mawb" db:"mawb"`
    HAWB                       string           `json:"hawb" db:"hawb"`
    ShipperNameAndAddress      string           `json:"shipperNameAndAddress" db:"shipper_name_and_address"`
    AWBIssuedBy                string           `json:"awbIssuedBy" db:"awb_issued_by"`
    ConsigneeNameAndAddress    string           `json:"consigneeNameAndAddress" db:"consignee_name_and_address"`
    IssuingCarrierAgentName    string           `json:"issuingCarrierAgentName" db:"issuing_carrier_agent_name"`
    AccountingInfomation       string           `json:"accountingInfomation" db:"accounting_infomation"`
    AgentsIATACode             string           `json:"agentsIATACode" db:"agents_iata_code"`
    AccountNo                  string           `json:"accountNo" db:"account_no"`
    AirportOfDeparture         string           `json:"airportOfDeparture" db:"airport_of_departure"`
    ReferenceNumber            string           `json:"referenceNumber" db:"reference_number"`
    OptionalShippingInfo1      string           `json:"optionalShippingInfo1" db:"optional_shipping_info1"`
    OptionalShippingInfo2      string           `json:"optionalShippingInfo2" db:"optional_shipping_info2"`
    RoutingTo                  string           `json:"routingTo" db:"routing_to"`
    RoutingBy                  string           `json:"routingBy" db:"routing_by"`
    DestinationTo1             string           `json:"destinationTo1" db:"destination_to1"`
    DestinationBy1             string           `json:"destinationBy1" db:"destination_by1"`
    DestinationTo2             string           `json:"destinationTo2" db:"destination_to2"`
    DestinationBy2             string           `json:"destinationBy2" db:"destination_by2"`
    Currency                   string           `json:"currency" db:"currency"`
    ChgsCode                   string           `json:"chgsCode" db:"chgs_code"`
    WtValPpd                   string           `json:"wtValPpd" db:"wt_val_ppd"`
    WtValColl                  string           `json:"wtValColl" db:"wt_val_coll"`
    OtherPpd                   string           `json:"otherPpd" db:"other_ppd"`
    OtherColl                  string           `json:"otherColl" db:"other_coll"`
    DeclaredValCarriage        string           `json:"declaredValCarriage" db:"declared_val_carriage"`
    DeclaredValCustoms         string           `json:"declaredValCustoms" db:"declared_val_customs"`
    AirportOfDestination       string           `json:"airportOfDestination" db:"airport_of_destination"`
    RequestedFlightDate1       string           `json:"requestedFlightDate1" db:"requested_flight_date1"`
    RequestedFlightDate2       string           `json:"requestedFlightDate2" db:"requested_flight_date2"`
    AmountOfInsurance          string           `json:"amountOfInsurance" db:"amount_of_insurance"`
    HandlingInfomation         string           `json:"handlingInfomation" db:"handling_infomation"`
    SCI                        string           `json:"sci" db:"sci"`
    Prepaid                    float64          `json:"prepaid" db:"prepaid"`
    ValuationCharge            float64          `json:"valuationCharge" db:"valuation_charge"`
    Tax                        float64          `json:"tax" db:"tax"`
    TotalOtherChargesDueAgent  float64          `json:"totalOtherChargesDueAgent" db:"total_other_charges_due_agent"`
    TotalOtherChargesDueCarrier float64         `json:"totalOtherChargesDueCarrier" db:"total_other_charges_due_carrier"`
    TotalPrepaid               float64          `json:"totalPrepaid" db:"total_prepaid"`
    CurrencyConversionRates    string           `json:"currencyConversionRates" db:"currency_conversion_rates"`
    Signature1                 string           `json:"signature1" db:"signature1"`
    Signature2Date             *time.Time       `json:"signature2Date" db:"signature2_date"`
    Signature2Place            string           `json:"signature2Place" db:"signature2_place"`
    Signature2Issuing          string           `json:"signature2Issuing" db:"signature2_issuing"`
    ShippingMark               string           `json:"shippingMark" db:"shipping_mark"`
    Status                     string           `json:"status" db:"status"`
    Items                      []DraftMAWBItem  `json:"items"`
    Charges                    []DraftMAWBCharge `json:"charges"`
    CreatedAt                  time.Time        `json:"createdAt" db:"created_at"`
    UpdatedAt                  time.Time        `json:"updatedAt" db:"updated_at"`
}

type DraftMAWBItem struct {
    ID                int                  `json:"id" db:"id"`
    DraftMAWBUUID     string               `json:"draft_mawb_uuid" db:"draft_mawb_uuid"`
    PiecesRCP         string               `json:"piecesRCP" db:"pieces_rcp"`
    GrossWeight       string               `json:"grossWeight" db:"gross_weight"`
    KgLb              string               `json:"kgLb" db:"kg_lb"`
    RateClass         string               `json:"rateClass" db:"rate_class"`
    TotalVolume       string               `json:"totalVolume" db:"total_volume"`
    ChargeableWeight  string               `json:"chargeableWeight" db:"chargeable_weight"`
    RateCharge        float64              `json:"rateCharge" db:"rate_charge"`
    Total             float64              `json:"total" db:"total"`
    NatureAndQuantity string               `json:"natureAndQuantity" db:"nature_and_quantity"`
    Dims              []DraftMAWBItemDim   `json:"dims"`
}

type DraftMAWBItemDim struct {
    ID               int    `json:"id" db:"id"`
    DraftMAWBItemID  int    `json:"draft_mawb_item_id" db:"draft_mawb_item_id"`
    Length           string `json:"length" db:"length"`
    Width            string `json:"width" db:"width"`
    Height           string `json:"height" db:"height"`
    Count            string `json:"count" db:"count"`
}

type DraftMAWBCharge struct {
    ID            int     `json:"id" db:"id"`
    DraftMAWBUUID string  `json:"draft_mawb_uuid" db:"draft_mawb_uuid"`
    Key           string  `json:"key" db:"charge_key"`
    Value         float64 `json:"value" db:"charge_value"`
}
```

## Additional API Endpoints for Draft MAWB

### Draft MAWB Status Management

```go
// POST /v1/mawbinfo/{uuid}/draft-mawb/confirm
func ConfirmDraftMAWB(c *gin.Context) {
    mawbUUID := c.Param("uuid")
    // Update status to 'Confirmed'
    // Return success response
}

// POST /v1/mawbinfo/{uuid}/draft-mawb/reject
func RejectDraftMAWB(c *gin.Context) {
    mawbUUID := c.Param("uuid")
    // Update status to 'Rejected'
    // Return success response
}

// GET /v1/mawbinfo/{uuid}/draft-mawb/print
func PrintDraftMAWB(c *gin.Context) {
    mawbUUID := c.Param("uuid")
    // Generate PDF from Draft MAWB data
    // Return PDF file
}
```

## Repository Pattern Examples for Draft MAWB

```go
type DraftMAWBRepository struct {
    db *sql.DB
}

func (r *DraftMAWBRepository) GetByMAWBUUID(mawbUUID string) (*DraftMAWB, error) {
    // Query draft_mawb table
    // Query related items, dims, and charges
    // Join and return complete structure
}

func (r *DraftMAWBRepository) CreateOrUpdate(draftMAWB *DraftMAWB) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Upsert draft_mawb record
    // Delete existing items, dims, charges
    // Insert new items with dims
    // Insert new charges

    return tx.Commit()
}

func (r *DraftMAWBRepository) UpdateStatus(uuid, status string) error {
    query := `UPDATE draft_mawb SET status = ?, updated_at = NOW() WHERE uuid = ?`
    _, err := r.db.Exec(query, status, uuid)
    return err
}
```

## Complex Calculations for Draft MAWB

### Volume and Chargeable Weight Calculation

```go
func CalculateVolumeAndChargeableWeight(dims []DraftMAWBItemDim, grossWeight string, kgLb string) (string, string) {
    var totalVolume float64

    for _, dim := range dims {
        length, _ := strconv.ParseFloat(dim.Length, 64)
        width, _ := strconv.ParseFloat(dim.Width, 64)
        height, _ := strconv.ParseFloat(dim.Height, 64)
        count, _ := strconv.ParseFloat(dim.Count, 64)

        if length > 0 && width > 0 && height > 0 && count > 0 {
            volume := (length * width * height) / 1000000 // cm³ to m³
            totalVolume += volume * count
        }
    }

    volumetricWeight := totalVolume * 166.67
    weight, _ := strconv.ParseFloat(grossWeight, 64)

    if kgLb == "lb" {
        weight = weight * 0.453592 // Convert lb to kg
    }

    chargeableWeight := math.Max(weight, volumetricWeight)

    return fmt.Sprintf("%.3f", totalVolume), fmt.Sprintf("%.2f", chargeableWeight)
}
```

### Financial Calculations

```go
func CalculateTotals(draftMAWB *DraftMAWB) {
    // Calculate total other charges due carrier
    var totalCharges float64
    for _, charge := range draftMAWB.Charges {
        totalCharges += charge.Value
    }

    var totalItemCharges float64
    for _, item := range draftMAWB.Items {
        totalItemCharges += item.RateCharge
    }

    draftMAWB.TotalOtherChargesDueCarrier = totalCharges + totalItemCharges

    // Calculate total prepaid
    draftMAWB.TotalPrepaid = draftMAWB.Prepaid +
                            draftMAWB.ValuationCharge +
                            draftMAWB.Tax +
                            draftMAWB.TotalOtherChargesDueAgent +
                            draftMAWB.TotalOtherChargesDueCarrier
}
```
