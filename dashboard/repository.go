package dashboard

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
)

type Repository interface {
	GetDashboardV1(ctx context.Context) (*DashboardV2Model, error)
}

type repository struct {
	contextTimeout time.Duration
}

func NewRepository(
	timeout time.Duration,
) Repository {
	return &repository{
		contextTimeout: timeout,
	}
}

func (r repository) GetDashboardV1(ctx context.Context) (*DashboardV2Model, error) {
	db := ctx.Value("postgreSQLConn").(*pg.DB)
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	summary := &SummaryModel{}
	_, err := db.QueryOneContext(ctx, pg.Scan(
		&summary.InboundTotalHawbCount,
		&summary.InboundTotalNetWeight,
		&summary.InboundTotalGrossWeight,
		&summary.OutboundTotalHawbCount,
		&summary.OutboundTotalNetWeight,
		&summary.OutboundTotalGrossWeight,
	), `
			WITH inbound AS (
				SELECT 
					COUNT(*) AS hawb_count,
					SUM(net_weight) AS net_weight,
					SUM(gross_weight) AS gross_weight
				FROM (
					SELECT 
						pimd.house_air_waybill,
						SUM(pimd.net_weight) AS net_weight,
						SUM(pimd.gross_weight) AS gross_weight
					FROM 
						tbl_pre_import_manifest_details pimd
					GROUP BY 
						pimd.house_air_waybill
				) sub_inbound
			),
			outbound AS (
				SELECT 
					COUNT(*) AS hawb_count,
					SUM(net_weight) AS net_weight,
					SUM(gross_weight) AS gross_weight
				FROM (
					SELECT 
						pemd.house_air_waybill,
						SUM(pemd.net_weight) AS net_weight,
						SUM(pemd.gross_weight) AS gross_weight
					FROM 
						tbl_pre_export_manifest_details pemd
					GROUP BY 
						pemd.house_air_waybill
				) sub_outbound
			)
			SELECT 
				inbound.hawb_count AS inbound_total_hawb_count,
				inbound.net_weight AS inbound_total_net_weight,
				inbound.gross_weight AS inbound_total_gross_weight,
				outbound.hawb_count AS outbound_total_hawb_count,
				outbound.net_weight AS outbound_total_net_weight,
				outbound.gross_weight AS outbound_total_gross_weight
			FROM 
				inbound
			CROSS JOIN
				outbound;
	 `)

	if err != nil {
		return nil, err
	}

	volumne := []*GetVolumneModel{}
	stmt, err := db.Prepare(`
		SELECT 
		tbl_tmp.mm_yy,
		(
			SELECT 
				count(distinct pimd.house_air_waybill) 
			from tbl_pre_import_manifest_details pimd
			where to_char(pimd.created_at, 'YYYY-MM') = tbl_tmp.mm_yy 
		) as inbound_total_hawb_count,
		(
			SELECT 
				count(distinct pemd.house_air_waybill) 
			from tbl_pre_export_manifest_details pemd
			where to_char(pemd.created_at, 'YYYY-MM') = tbl_tmp.mm_yy 
		) as outbound_total_hawb_count
	from (
	
		WITH months AS (SELECT * FROM generate_series(1, 12) AS t(n))

		SELECT 
		to_char( 
			TO_DATE(TO_CHAR(
				TO_DATE (months.n::text, 'MM'), 'Mon'
			) || ' ' || CURRENT_DATE, 'Mon YYYY')
			, 'YYYY-MM')as mm_yy
			
		FROM months
	) tbl_tmp
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.QueryContext(ctx, &volumne)
	if err != nil {
		return nil, err
	}

	volumneChart := &VolumneChartModel{}

	inboundVolumneChart := &VolumneChartDetailModel{Name: "Inbound"}
	outboundVolumneChart := &VolumneChartDetailModel{Name: "Outbound"}
	for _, v := range volumne {
		volumneChart.MmYy = append(volumneChart.MmYy, v.MmYy)
		inboundVolumneChart.Data = append(inboundVolumneChart.Data, v.InboundTotalHawbCount)
		outboundVolumneChart.Data = append(outboundVolumneChart.Data, v.OutboundTotalHawbCount)
	}
	volumneChart.Data = append(volumneChart.Data, inboundVolumneChart)
	volumneChart.Data = append(volumneChart.Data, outboundVolumneChart)

	return &DashboardV2Model{
		Summary: summary,
		Data:    volumneChart,
	}, nil
}
