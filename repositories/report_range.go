package repositories

import "time"

// bhutanTime is the report's day boundary (UTC+6, no daylight saving). A fixed zone avoids
// depending on tzdata in the container.
var bhutanTime = time.FixedZone("BTT", 6*60*60)

// reportRange turns the report's optional inclusive date range into half-open instants.
// activeFrom..to is the window for activity, newFrom..to for new registrations; with no range
// activity covers all time and new registrations the last 30 days.
type reportRange struct {
    activeFrom time.Time
    newFrom    time.Time
    to         time.Time
}

func newReportRange(fromDate, toDate *time.Time) reportRange {
    now := time.Now()
    r := reportRange{
        activeFrom: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
        newFrom:    now.AddDate(0, 0, -30),
        to:         now.Add(time.Minute),
    }
    if fromDate != nil {
        r.activeFrom = startOfDay(*fromDate)
        r.newFrom = r.activeFrom
    }
    if toDate != nil {
        r.to = startOfDay(*toDate).AddDate(0, 0, 1)
    }
    return r
}

// startOfDay is midnight in Bhutan of the calendar date the client sent. Dates arrive as
// midnight UTC ("2026-11-01" or "2026-11-01T00:00:00Z"), so the date is read in UTC.
func startOfDay(t time.Time) time.Time {
    y, m, d := t.UTC().Date()
    return time.Date(y, m, d, 0, 0, 0, 0, bhutanTime)
}
