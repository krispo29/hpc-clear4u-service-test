# Design Document

## Overview

The MAWB System Integration feature extends the existing MAWB Info system by adding two new major components: Cargo Manifest and Draft MAWB management. This design leverages the existing architecture patterns found in the codebase, including the factory pattern for services and repositories, Chi router for HTTP handling, and PostgreSQL with go-pg ORM for data persistence.

The system will create foreign key relationships between MAWB Info records and the new Cargo Manifest and Draft MAWB entities, enabling comprehensive cargo documentation management with proper data integrity and cascading operations.

## Architecture

### High-Level Architecture

The system follows the existing layered architecture pattern:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Layer    │    │   HTTP Layer    │    │   HTTP Layer    │
│ (Cargo Manifest)│    │ (Draft MAWB)    │    │  (MAWB Info)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Service Layer  │    │  Service Layer  │    │  Service Layer  │
│ (Cargo Manifest)│    │ (Draft MAWB)    │    │  (MAWB Info)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│Repository Layer │    │Repository Layer │    │Repository Layer │
│ (Cargo Manifest)│    │ (Draft MAWB)    │    │  (MAWB Info)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────────┐
                    │   PostgreSQL DB     │
                    │  - mawb_info        │
                    │  - cargo_manifest   │
                    │  - cargo_manifest_  │
                    │    items            │
                    │  - draft_mawb       │
                    │  - draft_mawb_items │
                    │  - draft_mawb_item_ │
                    │    dims             │
                    │  - draft_mawb_      │
                    │    charges          │
                    └─────────────────────┘
```

### Integration Points

The new components integrate with existing systems through:

1. **Factory Pattern Integration**: New services and repositories will be added to the existing ServiceFactory and RepositoryFactory
2. **Router Integration**: New handlers will be mounted as sub-routes under `/v1/mawbinfo/{uuid}/`
3. **Database Integration**: New tables will reference existing `mawb_info` table via foreign keys
4. **Authentication**: Leverages existing JWT authentication middleware

## Components and Interfaces

### 1. Cargo Manifest Component

#### Service Interface
```go
type CargoManifestService interface {
    GetCargoManifest(ctx context.Context, mawbUUID string) (*CargoManifestResponse, error)
    CreateOrUpdateCargoManifest(ctx context.Context, mawbUUID string, req *CargoManifestRequest) (*CargoManifestResponse, error)
    ConfirmCargoManifest(ctx context.Context, mawbUUID string) error
    RejectCargoManifest(ctx context.Context, mawbUUID string) error
    GenerateCargoManifestPDF(ctx context.Context, mawbUUID string) ([]byte, error)
}
```

#### Repository Interface
```go
type CargoManifestRepository interface {
    GetByMAWBUUID(ctx context.Context, mawbUUID string) (*CargoManifest, error)
    CreateOrUpdate(ctx context.Context, manifest *CargoManifest) error
    UpdateStatus(ctx context.Context, uuid, status string) error
    ValidateMAWBExists(ctx context.Context, mawbUUID string) error
}
```

#### HTTP Handler
```go
type cargoManifestHandler struct {
    service CargoManifestService
}

// Routes:
// GET    /v1/mawbinfo/{uuid}/cargo-manifest
// POST   /v1/mawbinfo/{uuid}/cargo-manifest  
// POST   /v1/mawbinfo/{uuid}/cargo-manifest/confirm
// POST   /v1/mawbinfo/{uuid}/cargo-manifest/reject
// GET    /v1/mawbinfo/{uuid}/cargo-manifest/print
```

### 2. Draft MAWB Component

#### Service Interface
```go
type DraftMAWBService interface {
    GetDraftMAWB(ctx context.Context, mawbUUID string) (*DraftMAWBResponse, error)
    CreateOrUpdateDraftMAWB(ctx context.Context, mawbUUID string, req *DraftMAWBRequest) (*DraftMAWBResponse, error)
    ConfirmDraftMAWB(ctx context.Context, mawbUUID string) error
    RejectDraftMAWB(ctx context.Context, mawbUUID string) error
    GenerateDraftMAWBPDF(ctx context.Context, mawbUUID string) ([]byte, error)
}
```

#### Repository Interface
```go
type DraftMAWBRepository interface {
    GetByMAWBUUID(ctx context.Context, mawbUUID string) (*DraftMAWB, error)
    CreateOrUpdate(ctx context.Context, draftMAWB *DraftMAWB) error
    UpdateStatus(ctx context.Context, uuid, status string) error
    ValidateMAWBExists(ctx context.Context, mawbUUID string) error
}
```

#### HTTP Handler
```go
type draftMAWBHandler struct {
    service DraftMAWBService
}

// Routes:
// GET    /v1/mawbinfo/{uuid}/draft-mawb
// POST   /v1/mawbinfo/{uuid}/draft-mawb
// POST   /v1/mawbinfo/{uuid}/draft-mawb/confirm
// POST   /v1/mawbinfo/{uuid}/draft-mawb/reject
// GET    /v1/mawbinfo/{uuid}/draft-mawb/print
```

### 3. Enhanced MAWB Info Handler

The existing MAWB Info handler will be extended to include sub-routes for cargo manifest and draft MAWB:

```go
func (h *mawbInfoHandler) router() chi.Router {
    r := chi.NewRouter()
    
    // Existing routes
    r.Post("/", h.createMawbInfo)
    r.Get("/", h.getAllMawbInfo)
    r.Get("/{uuid}", h.getMawbInfo)
    r.Put("/{uuid}", h.updateMawbInfo)
    r.Delete("/{uuid}", h.deleteMawbInfo)
    r.Delete("/{uuid}/attachments", h.deleteMawbInfoAttachment)
    
    // New sub-routes
    r.Route("/{uuid}", func(r chi.Router) {
        cargoManifestSvc := cargoManifestHandler{h.cargoManifestService}
        r.Mount("/cargo-manifest", cargoManifestSvc.router())
        
        draftMAWBSvc := draftMAWBHandler{h.draftMAWBService}
        r.Mount("/draft-mawb", draftMAWBSvc.router())
    })
    
    return r
}
```

## Data Models

### 1. Cargo Manifest Models

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
    CreatedAt               time.Time `json:"createdAt" db:"created_at"`
}
```

### 2. Draft MAWB Models

```go
type DraftMAWB struct {
    UUID                        string              `json:"uuid" db:"uuid"`
    MAWBInfoUUID               string              `json:"mawb_info_uuid" db:"mawb_info_uuid"`
    CustomerUUID               string              `json:"customerUUID" db:"customer_uuid"`
    AirlineLogo                string              `json:"airlineLogo" db:"airline_logo"`
    AirlineName                string              `json:"airlineName" db:"airline_name"`
    MAWB                       string              `json:"mawb" db:"mawb"`
    HAWB                       string              `json:"hawb" db:"hawb"`
    ShipperNameAndAddress      string              `json:"shipperNameAndAddress" db:"shipper_name_and_address"`
    ConsigneeNameAndAddress    string              `json:"consigneeNameAndAddress" db:"consignee_name_and_address"`
    // ... (additional 40+ fields as per requirements)
    Status                     string              `json:"status" db:"status"`
    Items                      []DraftMAWBItem     `json:"items"`
    Charges                    []DraftMAWBCharge   `json:"charges"`
    CreatedAt                  time.Time           `json:"createdAt" db:"created_at"`
    UpdatedAt                  time.Time           `json:"updatedAt" db:"updated_at"`
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
    CreatedAt         time.Time            `json:"createdAt" db:"created_at"`
}

type DraftMAWBItemDim struct {
    ID               int       `json:"id" db:"id"`
    DraftMAWBItemID  int       `json:"draft_mawb_item_id" db:"draft_mawb_item_id"`
    Length           string    `json:"length" db:"length"`
    Width            string    `json:"width" db:"width"`
    Height           string    `json:"height" db:"height"`
    Count            string    `json:"count" db:"count"`
    CreatedAt        time.Time `json:"createdAt" db:"created_at"`
}

type DraftMAWBCharge struct {
    ID            int       `json:"id" db:"id"`
    DraftMAWBUUID string    `json:"draft_mawb_uuid" db:"draft_mawb_uuid"`
    Key           string    `json:"key" db:"charge_key"`
    Value         float64   `json:"value" db:"charge_value"`
    CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}
```

### 3. Request/Response Models

```go
// Cargo Manifest
type CargoManifestRequest struct {
    MAWBNumber      string                     `json:"mawbNumber" validate:"required"`
    PortOfDischarge string                     `json:"portOfDischarge"`
    FlightNo        string                     `json:"flightNo"`
    FreightDate     string                     `json:"freightDate"`
    Shipper         string                     `json:"shipper"`
    Consignee       string                     `json:"consignee"`
    TotalCtn        string                     `json:"totalCtn"`
    Transshipment   string                     `json:"transshipment"`
    Items           []CargoManifestItemRequest `json:"items"`
}

type CargoManifestItemRequest struct {
    HAWBNo                  string `json:"hawbNo"`
    Pkgs                    string `json:"pkgs"`
    GrossWeight             string `json:"grossWeight"`
    Destination             string `json:"dst"`
    Commodity               string `json:"commodity"`
    ShipperNameAndAddress   string `json:"shipperNameAndAddress"`
    ConsigneeNameAndAddress string `json:"consigneeNameAndAddress"`
}

// Draft MAWB
type DraftMAWBRequest struct {
    CustomerUUID            string                  `json:"customerUUID"`
    AirlineLogo            string                  `json:"airlineLogo"`
    AirlineName            string                  `json:"airlineName"`
    MAWB                   string                  `json:"mawb" validate:"required"`
    HAWB                   string                  `json:"hawb"`
    ShipperNameAndAddress  string                  `json:"shipperNameAndAddress"`
    ConsigneeNameAndAddress string                 `json:"consigneeNameAndAddress"`
    // ... (additional fields)
    Items                  []DraftMAWBItemRequest  `json:"items"`
    Charges                []DraftMAWBChargeRequest `json:"charges"`
}

type DraftMAWBItemRequest struct {
    PiecesRCP         string                    `json:"piecesRCP"`
    GrossWeight       string                    `json:"grossWeight"`
    KgLb              string                    `json:"kgLb"`
    RateClass         string                    `json:"rateClass"`
    RateCharge        float64                   `json:"rateCharge"`
    NatureAndQuantity string                    `json:"natureAndQuantity"`
    Dims              []DraftMAWBItemDimRequest `json:"dims"`
}

type DraftMAWBItemDimRequest struct {
    Length string `json:"length"`
    Width  string `json:"width"`
    Height string `json:"height"`
    Count  string `json:"count"`
}

type DraftMAWBChargeRequest struct {
    Key   string  `json:"key" validate:"required"`
    Value float64 `json:"value"`
}
```

## Error Handling

### Error Types

The system will implement consistent error handling following the existing patterns:

```go
// Custom error types
type MAWBNotFoundError struct {
    UUID string
}

func (e MAWBNotFoundError) Error() string {
    return fmt.Sprintf("MAWB Info not found: %s", e.UUID)
}

type CargoManifestNotFoundError struct {
    MAWBUUID string
}

func (e CargoManifestNotFoundError) Error() string {
    return fmt.Sprintf("Cargo Manifest not found for MAWB: %s", e.MAWBUUID)
}

type DraftMAWBNotFoundError struct {
    MAWBUUID string
}

func (e DraftMAWBNotFoundError) Error() string {
    return fmt.Sprintf("Draft MAWB not found for MAWB: %s", e.MAWBUUID)
}
```

### Error Response Mapping

```go
func mapErrorToHTTPResponse(err error) render.Renderer {
    switch e := err.(type) {
    case MAWBNotFoundError, CargoManifestNotFoundError, DraftMAWBNotFoundError:
        return &ErrResponse{
            Err:            err,
            HTTPStatusCode: http.StatusNotFound,
            AppCode:        constant.CodeError,
            Message:        e.Error(),
        }
    case ValidationError:
        return &ErrResponse{
            Err:            err,
            HTTPStatusCode: http.StatusBadRequest,
            AppCode:        constant.CodeError,
            Message:        e.Error(),
        }
    default:
        return &ErrResponse{
            Err:            err,
            HTTPStatusCode: http.StatusInternalServerError,
            AppCode:        constant.CodeError,
            Message:        "Internal server error",
        }
    }
}
```

## Testing Strategy

### 1. Unit Testing

**Service Layer Tests:**
- Test business logic validation
- Test calculation functions (volume, chargeable weight, financial totals)
- Test error handling scenarios
- Mock repository dependencies

**Repository Layer Tests:**
- Test database operations with test database
- Test transaction handling
- Test foreign key constraint validation
- Test cascading delete operations

**Handler Layer Tests:**
- Test HTTP request/response handling
- Test request validation
- Test authentication middleware integration
- Mock service dependencies

### 2. Integration Testing

**Database Integration:**
- Test complete CRUD operations
- Test foreign key relationships
- Test transaction rollback scenarios
- Test concurrent access patterns

**API Integration:**
- Test complete request/response cycles
- Test authentication flows
- Test error response formats
- Test PDF generation endpoints

### 3. Test Data Management

```go
// Test fixtures for consistent test data
type TestFixtures struct {
    MAWBInfo        *mawbinfo.MawbInfoResponse
    CargoManifest   *CargoManifest
    DraftMAWB       *DraftMAWB
}

func SetupTestFixtures(db *pg.DB) *TestFixtures {
    // Create test MAWB Info
    // Create test Cargo Manifest
    // Create test Draft MAWB
    // Return fixtures
}

func CleanupTestFixtures(db *pg.DB, fixtures *TestFixtures) {
    // Clean up test data in reverse dependency order
}
```

## Performance Considerations

### 1. Database Optimization

**Indexing Strategy:**
```sql
-- Foreign key indexes for join performance
CREATE INDEX idx_cargo_manifest_mawb_info_uuid ON cargo_manifest(mawb_info_uuid);
CREATE INDEX idx_cargo_manifest_items_cargo_manifest_uuid ON cargo_manifest_items(cargo_manifest_uuid);
CREATE INDEX idx_draft_mawb_mawb_info_uuid ON draft_mawb(mawb_info_uuid);
CREATE INDEX idx_draft_mawb_items_draft_mawb_uuid ON draft_mawb_items(draft_mawb_uuid);
CREATE INDEX idx_draft_mawb_item_dims_draft_mawb_item_id ON draft_mawb_item_dims(draft_mawb_item_id);
CREATE INDEX idx_draft_mawb_charges_draft_mawb_uuid ON draft_mawb_charges(draft_mawb_uuid);

-- Status-based queries
CREATE INDEX idx_cargo_manifest_status ON cargo_manifest(status);
CREATE INDEX idx_draft_mawb_status ON draft_mawb(status);

-- Date-based queries
CREATE INDEX idx_cargo_manifest_created_at ON cargo_manifest(created_at);
CREATE INDEX idx_draft_mawb_created_at ON draft_mawb(created_at);
```

**Query Optimization:**
- Use prepared statements for repeated queries
- Implement efficient JOIN strategies for nested data retrieval
- Use transactions for multi-table operations
- Implement connection pooling (already handled by go-pg)

### 2. Memory Management

**Large Dataset Handling:**
- Implement pagination for list endpoints
- Use streaming for PDF generation
- Limit nested data depth in responses
- Implement request size limits

**Caching Strategy:**
- Cache frequently accessed MAWB Info records
- Cache PDF templates and fonts
- Implement Redis caching for session data
- Use HTTP caching headers for static resources

### 3. Calculation Performance

**Optimization for Complex Calculations:**
```go
// Pre-calculate and cache expensive computations
type CalculationCache struct {
    VolumetricWeight map[string]float64
    ChargeableTotals map[string]float64
    mutex           sync.RWMutex
}

func (c *CalculationCache) GetOrCalculateVolumetricWeight(dims []DraftMAWBItemDim) float64 {
    key := generateDimensionKey(dims)
    
    c.mutex.RLock()
    if weight, exists := c.VolumetricWeight[key]; exists {
        c.mutex.RUnlock()
        return weight
    }
    c.mutex.RUnlock()
    
    weight := calculateVolumetricWeight(dims)
    
    c.mutex.Lock()
    c.VolumetricWeight[key] = weight
    c.mutex.Unlock()
    
    return weight
}
```

## Security Considerations

### 1. Authentication & Authorization

- Leverage existing JWT authentication middleware
- Validate user permissions for MAWB Info access
- Implement role-based access control for status changes
- Audit trail for sensitive operations

### 2. Input Validation

```go
// Comprehensive input validation
func validateCargoManifestRequest(req *CargoManifestRequest) error {
    if err := validator.New().Struct(req); err != nil {
        return err
    }
    
    // Business logic validation
    if len(req.Items) > MaxCargoManifestItems {
        return errors.New("too many cargo manifest items")
    }
    
    // Sanitize string inputs
    req.MAWBNumber = sanitizeInput(req.MAWBNumber)
    req.Shipper = sanitizeInput(req.Shipper)
    req.Consignee = sanitizeInput(req.Consignee)
    
    return nil
}
```

### 3. SQL Injection Prevention

- Use parameterized queries (already implemented with go-pg)
- Validate UUID formats
- Sanitize user inputs
- Implement query timeouts

### 4. Data Privacy

- Mask sensitive information in logs
- Implement data retention policies
- Secure PDF generation process
- Encrypt sensitive data at rest

## PDF Generation Strategy

### 1. PDF Library Selection

Use existing PDF generation patterns in the codebase with enhanced templates:

```go
type PDFGenerator struct {
    templates map[string]*template.Template
    fonts     map[string][]byte
}

func NewPDFGenerator() *PDFGenerator {
    return &PDFGenerator{
        templates: make(map[string]*template.Template),
        fonts:     loadFonts(),
    }
}

func (g *PDFGenerator) GenerateCargoManifestPDF(manifest *CargoManifest) ([]byte, error) {
    // Use existing font loading patterns
    // Create PDF with cargo manifest template
    // Return PDF bytes
}

func (g *PDFGenerator) GenerateDraftMAWBPDF(draftMAWB *DraftMAWB) ([]byte, error) {
    // Use existing font loading patterns  
    // Create PDF with draft MAWB template
    // Return PDF bytes
}
```

### 2. Template Management

- Create reusable PDF templates for cargo manifest and draft MAWB
- Support multiple languages (Thai/English) following existing patterns
- Implement template caching for performance
- Support custom branding/logos

### 3. Error Handling

- Graceful degradation for PDF generation failures
- Retry mechanisms for temporary failures
- Detailed error logging for debugging
- Alternative formats (HTML) as fallback