export function createApexAreaChart(id, title, seriesNames, stacked = false) {
    const options = {
        chart: {
            type: "area",
            height: 250,
            zoom: {
                type: "x",
                enabled: true,
                autoScaleYaxis: true
            },
            toolbar: {
                autoSelected: "zoom"
            },
            stacked: stacked,
            animations: {
                enabled: true
            }
        },
        stroke: {
            curve: seriesNames.length > 1 ? "straight" : "smooth",
            width: 2
        },
        fill: {
            type: "gradient",
            gradient: {
                shadeIntensity: 1,
                opacityFrom: 0.4,
                opacityTo: 0,
                stops: [0, 90, 100]
            }
        },
        dataLabels: {
            enabled: false
        },
        markers: {
            size: 0
        },
        title: {
            text: title,
            align: "left",
            style: {
                fontSize: "14px",
                fontWeight: 600,
                color: "#263238"
            }
        },
        xaxis: {
            type: "datetime",
            labels: {
                datetimeFormatter: {
                    month: "MMM 'yy",
                    day: "dd MMM",
                    hour: "HH:mm",
                    minute: "HH:mm"
                }
            }
        },
        yaxis: {
            labels: {
                formatter: val => val.toFixed(2)
            },
            title: {
                text: "Value"
            }
        },
        tooltip: {
            shared: true,
            intersect: false,
            x: { format: "MMM dd HH:mm" },
            y: { formatter: val => val.toFixed(2) }
        },
        series: seriesNames.map(name => ({ name, data: [] }))
    };

    const chart = new ApexCharts(document.getElementById(id), options);
    chart.render();
    return chart;
}

export function createApexDonutChart(id, title) {
    const options = {
        chart: {
            type: "donut",
            height: 250
        },
        labels: ["User", "System", "Idle", "Other"],
        series: [0, 0, 0, 0],
        title: {
            text: title,
            align: "left",
            style: {
                fontSize: "14px",
                fontWeight: 600,
                color: "#263238"
            }
        },
        tooltip: {
            y: { formatter: val => val.toFixed(1) }
        },
        legend: {
            position: "bottom"
        }
    };

    const chart = new ApexCharts(document.getElementById(id), options);
    chart.render();
    return chart;
}
