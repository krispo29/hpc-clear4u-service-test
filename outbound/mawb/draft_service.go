package outbound

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"

	"github.com/jung-kurt/gofpdf"
)

func (s *service) GetAllMawbDraft(ctx context.Context, start, end string) ([]*GetAllMawbDraftModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if resp, err := s.selfRepo.GetAllMawbDraft(ctx, start, end); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (s *service) GetOneMawbDraft(ctx context.Context, uuid string) (*GetMawbDraftModel, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if resp, err := s.selfRepo.GetOneMawbDraft(ctx, uuid); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (s *service) PrintMawbDraft(ctx context.Context, uuid string) (bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if resp, err := s.selfRepo.GetOneMawbDraft(ctx, uuid); err != nil {
		return bytes.Buffer{}, err
	} else {
		return s.generateDraftMawb(ctx, &resp.RequestDraftModel, false)
	}
}

func (s *service) PreviewDraftMawb(ctx context.Context, data *RequestDraftModel) (bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	return s.generateDraftMawb(ctx, data, true)

}
func (s *service) SaveDraftMawb(ctx context.Context, data *RequestDraftModel) (bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	if err := s.selfRepo.CreateMawbDraft(ctx, data); err != nil {
		return bytes.Buffer{}, err
	}

	return s.generateDraftMawb(ctx, data, false)
}

func (s *service) UpdateDraftMawb(ctx context.Context, data *RequestUpdateMawbDraftModel) (bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	log.Println(data.UUID)

	if err := s.selfRepo.UpdateMawbDraft(ctx, data); err != nil {
		return bytes.Buffer{}, err
	}

	return s.generateDraftMawb(ctx, &data.RequestDraftModel, false)
}

func (s *service) generateDraftMawb(ctx context.Context, data *RequestDraftModel, isPreview bool) (bytes.Buffer, error) {
	// Loading Font
	frontTHSarabunNew, err := ioutil.ReadFile("assets/THSarabunNew.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewBold, err := ioutil.ReadFile("assets/THSarabunNew Bold.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewBoldItalic, err := ioutil.ReadFile("assets/THSarabunNew BoldItalic.ttf")
	if err != nil {
		log.Println(err)
	}

	frontTHSarabunNewItalic, err := ioutil.ReadFile("assets/THSarabunNew Italic.ttf")
	if err != nil {
		log.Println(err)
	}

	var buf bytes.Buffer

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		SizeStr:        gofpdf.PageSizeA4,
	})

	pdf.AddUTF8FontFromBytes("THSarabunNew", "", frontTHSarabunNew)
	pdf.AddUTF8FontFromBytes("THSarabunNew Bold", "", frontTHSarabunNewBold)
	pdf.AddUTF8FontFromBytes("THSarabunNew BoldItalic", "", frontTHSarabunNewBoldItalic)
	pdf.AddUTF8FontFromBytes("THSarabunNew Italic", "", frontTHSarabunNewItalic)

	pdf.SetFont("THSarabunNew Bold", "", 10)
	for i1 := 0; i1 < 1; i1++ {

		pdf.AddPage()
		pdf.SetMargins(0, 0, 0)
		pdf.SetAutoPageBreak(true, 0)
		width, height := pdf.GetPageSize()

		if isPreview {
			pdf.ImageOptions(
				"assets/bg-mawb.png", // path to the image
				0, 0,                 // x, y positions
				width, height, // width, height to fit full page
				false, // do not flow the image
				gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
				0, // link
				"",
			)
		}

		pdf.SetFont("THSarabunNew Bold", "", 18)
		pdf.SetXY(7, 4)
		pdf.MultiCell(51, 5, data.Mawb, "0", "L", false)

		pdf.SetFont("THSarabunNew Bold", "", 18)
		pdf.SetXY(143, 4)
		pdf.MultiCell(51, 5, data.Hawb, "0", "C", false)

		pdf.SetFont("THSarabunNew Bold", "", 10)

		pdf.SetXY(9, 19)
		pdf.MultiCell(89, 3, data.ShipperNameAndAddress, "0", "LT", false)

		pdf.SetFont("THSarabunNew Bold", "", 18)
		pdf.SetXY(118, 13)
		pdf.MultiCell(77, 20, data.AwbIssuedBy, "0", "C", false)

		pdf.SetFont("THSarabunNew Bold", "", 10)
		pdf.SetXY(9, 45)
		pdf.MultiCell(89, 3, data.ConsigneeNameAndAddress, "0", "LT", false)

		pdf.SetFont("THSarabunNew Bold", "", 14)
		pdf.SetXY(8, 65)
		pdf.CellFormat(89, 15, data.IssuingCarrierAgentName, "0", 0, "C", false, 0, "")

		pdf.SetFont("THSarabunNew Bold", "", 18)
		pdf.SetXY(98, 67)
		pdf.MultiCell(97, 6, data.AccountingInfomation, "0", "C", false)

		pdf.SetFont("THSarabunNew Bold", "", 10)
		pdf.SetXY(8, 84)
		pdf.CellFormat(44, 6, data.AgentsIATACode, "0", 0, "L", false, 0, "")

		pdf.SetXY(53, 84)
		pdf.MultiCell(44, 6, data.AccountNo, "0", "L", false)

		pdf.SetXY(8, 92)
		pdf.MultiCell(89, 6, data.AirportOfDeparture, "0", "C", false)

		pdf.SetXY(97, 92)
		pdf.MultiCell(35, 6, data.ReferenceNumber, "0", "C", false)

		pdf.SetXY(132, 92)
		pdf.MultiCell(30, 6, data.OptionalShippingInfo1, "0", "C", false)

		pdf.SetXY(162, 92)
		pdf.MultiCell(32, 6, data.OptionalShippingInfo2, "0", "C", false)

		pdf.SetXY(8, 102)
		pdf.CellFormat(11, 6, data.RoutingTo, "0", 0, "C", false, 0, "")

		pdf.SetXY(20, 102)
		pdf.CellFormat(40, 6, data.RoutingBy, "0", 0, "C", false, 0, "")

		pdf.SetXY(62, 102)
		pdf.CellFormat(11, 6, data.DestinationTo1, "0", 0, "C", false, 0, "")

		pdf.SetXY(73, 102)
		pdf.CellFormat(8, 6, data.DestinationBy1, "0", 0, "C", false, 0, "")

		pdf.SetXY(81, 102)
		pdf.CellFormat(9, 6, data.DestinationTo2, "0", 0, "C", false, 0, "")

		pdf.SetXY(90, 102)
		pdf.CellFormat(8, 6, data.DestinationBy2, "0", 0, "C", false, 0, "")

		pdf.SetXY(98, 102)
		pdf.CellFormat(10, 6, data.Currency, "0", 0, "C", false, 0, "")

		pdf.SetXY(106, 102)
		pdf.CellFormat(8, 6, data.ChgsCode, "0", 0, "C", false, 0, "")

		pdf.SetXY(112, 102)
		pdf.CellFormat(8, 6, data.WtValPpd, "0", 0, "C", false, 0, "")

		pdf.SetXY(117, 102)
		pdf.CellFormat(8, 6, data.WtValColl, "0", 0, "C", false, 0, "")

		pdf.SetXY(122, 102)
		pdf.CellFormat(8, 6, data.OtherPpd, "0", 0, "C", false, 0, "")

		pdf.SetXY(127, 102)
		pdf.CellFormat(8, 6, data.OtherColl, "0", 0, "C", false, 0, "")

		pdf.SetXY(132, 102)
		pdf.CellFormat(31, 6, data.DeclaredValCarriage, "0", 0, "C", false, 0, "")

		pdf.SetXY(163, 102)
		pdf.CellFormat(31, 6, data.DeclaredValCustoms, "0", 0, "C", false, 0, "")

		pdf.SetXY(9, 111)
		pdf.CellFormat(45, 7, data.AirportOfDestination, "0", 0, "L", false, 0, "")

		pdf.SetXY(54, 111)
		pdf.MultiCell(22, 7, data.RequestedFlightDate1, "0", "C", false)

		pdf.SetXY(75, 111)
		pdf.MultiCell(22, 7, data.RequestedFlightDate2, "0", "C", false)

		pdf.SetXY(98, 111)
		pdf.MultiCell(28, 7, data.AmountOfInsurance, "0", "C", false)

		pdf.SetFont("THSarabunNew Bold", "", 14)
		pdf.SetXY(9, 119)
		pdf.MultiCell(156, 12, data.HandlingInfomation, "0", "L", false)

		pdf.SetXY(165, 125)
		pdf.MultiCell(30, 7, data.Sci, "0", "L", false)

		pdf.SetFont("THSarabunNew Bold", "", 10)
		dstartX := float64(8)
		dStartY := float64(143)
		pdf.SetY(dStartY)
		for _, v := range data.Items {
			pdf.SetX(dstartX)
			pdf.CellFormat(10, 7, v.PiecesRCP, "0", 0, "CM", false, 0, "")
			pdf.CellFormat(19, 7, v.GrossWeight, "0", 0, "CM", false, 0, "")
			pdf.CellFormat(25, 7, v.RateClass, "0", 0, "L", false, 0, "")         // Add
			pdf.CellFormat(19, 7, v.ChargeableWeight, "0", 0, "CM", false, 0, "") // Add
			pdf.CellFormat(22, 7, v.RateCharge, "0", 0, "CM", false, 0, "")       // Add
			pdf.CellFormat(35, 7, v.Total, "0", 0, "CM", false, 0, "")            // Add
			pdf.CellFormat(3, 7, "", "0", 0, "CM", false, 0, "")                  // Add
			pdf.MultiCell(57, 4, v.NatureAndQuantity, "0", "L", false)
			pdf.Ln(2.5)
		}

		pdf.SetXY(6, 204)
		pdf.MultiCell(33, 6, data.Prepaid, "0", "C", false) // Add

		pdf.SetXY(6, 212)
		pdf.MultiCell(33, 6, data.ValuationCharge, "0", "C", false) // Add

		pdf.SetXY(6, 221)
		pdf.MultiCell(33, 6, data.Tax, "0", "C", false) // Add

		pdf.SetXY(6, 230)
		pdf.MultiCell(33, 6, data.TotalOtherChargesDueAgent, "0", "C", false) // Add

		pdf.SetXY(6, 239)
		pdf.MultiCell(33, 6, data.TotalOtherChargesDueCarrier, "0", "C", false) // Add

		pdf.SetXY(6, 257)
		pdf.MultiCell(33, 6, data.TotalPrepaid, "0", "C", false) // Add

		pdf.SetXY(6, 266)
		pdf.MultiCell(33, 6, data.CurrencyConversionRates, "0", "C", false) // Add

		// Other Charge
		pdf.SetY(208)
		otherServices := []*OtherServiceModel{
			{data.TerminalChargeKey, data.TerminalChargeVal},
			{data.MrKey, data.MrVal},
			{data.BcKey, data.BcVal},
			{data.AweFeeKey, data.AweFeeVal},
			{data.CcKey, data.CcVal}, // Add
		}

		for _, v := range otherServices {
			pdf.SetX(83)
			pdf.CellFormat(45, 3.6, v.Key, "0", 0, "L", false, 0, "")
			pdf.MultiCell(20, 3.6, v.Value, "0", "L", false)
			pdf.Ln(1)
		}

		pdf.SetFont("THSarabunNew Bold", "", 14)
		pdf.SetXY(82, 247)
		pdf.MultiCell(120, 5, data.Signature1, "0", "C", false)

		// pdf.SetXY(85, 266)
		// pdf.MultiCell(120, 5, data.Signature2, "0", "C", false)
		pdf.SetXY(82, 265)
		pdf.CellFormat(39, 5, data.Signature2Date, "0", 0, "C", false, 0, "")
		pdf.CellFormat(39, 5, data.Signature2Place, "0", 0, "C", false, 0, "")   // Add
		pdf.CellFormat(39, 5, data.Signature2Issuing, "0", 0, "C", false, 0, "") // Add

		pdf.SetFont("THSarabunNew Bold", "", 18)
		pdf.SetXY(136, 280)
		pdf.MultiCell(51, 3, data.Mawb, "0", "L", false)

		if isPreview {
			// Add watermark
			pdf.SetFont("THSarabunNew Bold", "", 78)
			pdf.SetTextColor(200, 200, 200) // Light grey
			pdf.TransformBegin()
			pdf.TransformRotate(45, width/2, height/2) // Rotate 45 degrees
			// pdf.Text(width/2, 20, "1P r e v i e w")
			pdf.Text(width/2-40, height/2, "P R E V I E W")
			// pdf.Text(width/2-40, height/3, "P r e v i e w")
			pdf.TransformEnd()
		}

	}

	err = pdf.Output(&buf)
	pdf.Close()

	if err == nil {
		return buf, nil
	}

	return bytes.Buffer{}, err
}
