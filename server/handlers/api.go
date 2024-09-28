package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Gazer/pocketfunctions/models"
	"github.com/gin-gonic/gin"
)

var headerTemplate = `
	<small class="text-muted mb-1">%s</small>
 	<h3 class="card-title mb-0">%d</h3>
   <p class="small text-muted mb-0">
    <span class="fe fe-arrow-%s fe-12 text-%s">
    </span>
    <span>
    	%.1f%% Last week
    </span>
  </p>`

type PocketAPI struct {
	Db     *sql.DB
	Router *gin.Engine
}

func New() *PocketAPI {
	var api PocketAPI
	api.Db = models.InitDB()
	api.Router = gin.Default()
	// REST API
	api.registerApi()
	// Catch-all to execute functions if any match exists
	api.Router.NoRoute(api.Execute())
	return &api
}

func (api *PocketAPI) InitAdminUI() {
	api.Router.Static("/_/", "./public")
}

func (api *PocketAPI) registerApi() {
	apiGroup := api.Router.Group("/api")

	apiGroup.GET("/functions", api.getFunctionsHandler())
	apiGroup.GET("/functions/runs", api.getRunsHandler())
	apiGroup.GET("/functions/errors", api.getErrorsHandler())
	apiGroup.GET("/functions/avg", api.getAvgTimeHandler())
	apiGroup.GET("/functions/histogram", api.getHistogramHandler())
	apiGroup.POST("/create", Create(api.Db))
	apiGroup.POST("/upload/:id", Upload(api.Db))
}

func (api *PocketAPI) getFunctionsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		functions, err := models.GetFunctions(api.Db)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		var buffer strings.Builder
		for _, f := range functions {
			buffer.Write([]byte("<tr>"))
			writeTd(&buffer, strconv.Itoa(f.Id))
			writeTd(&buffer, string(f.Code))
			writeTd(&buffer, string(f.Uri))
			writeTd(&buffer, strconv.Itoa(f.Execution))
			writeTd(&buffer, strconv.FormatFloat(f.Average, 'f', 1, 64))
			buffer.Write([]byte("</tr>"))
		}
		c.String(http.StatusOK, buffer.String())
	}
}

func (api *PocketAPI) getHistogramHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, _ := models.GetHistogram(api.Db)
		var values []string
		var dates []string
		for _, v := range data {
			values = append(values, strconv.Itoa(v.Second))
			dates = append(dates, fmt.Sprintf("\"%s\"", v.First))
		}

		jsValues := strings.Join(values, ", ")
		jsDates := strings.Join(dates, ", ")

		js := fmt.Sprintf(`
			var lineChart,
  lineChartoptions = {
    series: [
      {
        name: "Functions",
        data: [
          %s
        ],
      },
    ],
    chart: {
      height: 350,
      type: "line",
      background: !1,
      zoom: {
        enabled: !1,
      },
      toolbar: {
        show: !1,
      },
    },
    theme: {
      mode: colors.chartTheme,
    },
    stroke: {
      show: !0,
      curve: "smooth",
      lineCap: "round",
      colors: chartColors,
      width: [3, 2, 3],
      dashArray: [0, 0, 0],
    },
    dataLabels: {
      enabled: !1,
    },
    responsive: [
      {
        breakpoint: 480,
        options: {
          legend: {
            position: "bottom",
            offsetX: -10,
            offsetY: 0,
          },
        },
      },
    ],
    markers: {
      size: 4,
      colors: base.primaryColor,
      strokeColors: colors.borderColor,
      strokeWidth: 2,
      strokeOpacity: 0.9,
      strokeDashArray: 0,
      fillOpacity: 1,
      discrete: [],
      shape: "circle",
      radius: 2,
      offsetX: 0,
      offsetY: 0,
      onClick: void 0,
      onDblClick: void 0,
      showNullDataPoints: !0,
      hover: {
        size: void 0,
        sizeOffset: 3,
      },
    },
    xaxis: {
      type: "datetime",
      categories: [
      	%s
      ],
      labels: {
        show: !0,
        trim: !1,
        minHeight: void 0,
        maxHeight: 120,
        style: {
          colors: colors.mutedColor,
          cssClass: "text-muted",
          fontFamily: base.defaultFontFamily,
        },
      },
      axisBorder: {
        show: !1,
      },
    },
    yaxis: {
      labels: {
        show: !0,
        trim: !1,
        offsetX: -10,
        minHeight: void 0,
        maxHeight: 120,
        style: {
          colors: colors.mutedColor,
          cssClass: "text-muted",
          fontFamily: base.defaultFontFamily,
        },
      },
    },
    legend: {
      position: "top",
      fontFamily: base.defaultFontFamily,
      fontWeight: 400,
      labels: {
        colors: colors.mutedColor,
        useSeriesColors: !1,
      },
      markers: {
        width: 10,
        height: 10,
        strokeWidth: 0,
        strokeColor: colors.borderColor,
        fillColors: chartColors,
        radius: 6,
        customHTML: void 0,
        onClick: void 0,
        offsetX: 0,
        offsetY: 0,
      },
      itemMargin: {
        horizontal: 10,
        vertical: 0,
      },
      onItemClick: {
        toggleDataSeries: !0,
      },
      onItemHover: {
        highlightDataSeries: !0,
      },
    },
    grid: {
      show: !0,
      borderColor: colors.borderColor,
      strokeDashArray: 0,
      position: "back",
      xaxis: {
        lines: {
          show: !1,
        },
      },
      yaxis: {
        lines: {
          show: !0,
        },
      },
      row: {
        colors: void 0,
        opacity: 0.5,
      },
      column: {
        colors: void 0,
        opacity: 0.5,
      },
      padding: {
        top: 0,
        right: 0,
        bottom: 0,
        left: 0,
      },
    },
  },
  lineChartCtn = document.querySelector("#lineChart");
  lineChartCtn && (lineChart = new ApexCharts(lineChartCtn, lineChartoptions)).render();
`, jsValues, jsDates)

		c.String(http.StatusOK, js)
	}
}

func (api *PocketAPI) getRunsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := models.GetTotalCalls(api.Db)
		var response string
		if err != nil {
			response = err.Error()
		} else {
			response = headerSection("Total Runs", data.First, float64(data.Second))
		}
		c.String(http.StatusOK, response)
	}
}

func (api *PocketAPI) getErrorsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := models.GetTotalErrors(api.Db)
		var response string
		if err != nil {
			response = err.Error()
		} else {
			response = headerSection("Errors", data.First, float64(data.Second))
		}
		c.String(http.StatusOK, response)
	}
}

func (api *PocketAPI) getAvgTimeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := models.GetAvgTime(api.Db)
		var response string
		if err != nil {
			response = err.Error()
		} else {
			response = headerSection("Avg Time (ms)", data.First, float64(data.Second))
		}
		c.String(http.StatusOK, response)
	}
}

func headerSection(title string, total int, variation float64) string {
	var arrow string
	var color string
	if variation >= 0 {
		arrow = "up"
		color = "success"
	} else {
		arrow = "down"
		color = "danger"
	}
	return fmt.Sprintf(headerTemplate, title, total, arrow, color, variation)
}

func writeTd(buffer *strings.Builder, value string) {
	buffer.Write([]byte("<td>"))
	buffer.Write([]byte(value))
	buffer.Write([]byte("</td>"))
}
