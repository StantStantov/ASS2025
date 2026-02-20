package components

import (
	"StantStantov/ASS/internal/simulation"
	"StantStantov/ASS/internal/simulation/metrics"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/StantStantov/rps/swamp/atomic"
	"github.com/StantStantov/rps/swamp/collections/sparsemap"
)

func DrawTable() {
	allAlertsAtomic := &simulation.MetricsSystem.Metrics[metrics.AlertsBufferedCounter]
	rewrittenAlertsAtomic := &simulation.MetricsSystem.Metrics[metrics.AlertsRewrittenCounter]
	allAlerts := atomic.LoadUint64(allAlertsAtomic)
	rewrittenAlerts := atomic.LoadUint64(rewrittenAlertsAtomic)
	rewritePercentage := float64(0)
	if rewrittenAlerts != 0 {
		rewritePercentage = float64(rewrittenAlerts) / float64(allAlerts)
	}

	addedJobsAtomic := &simulation.MetricsSystem.Metrics[metrics.JobsPendingCounter]
	skippedJobsAtomic := &simulation.MetricsSystem.Metrics[metrics.JobsSkippedCounter]
	addedJobs := atomic.LoadUint64(addedJobsAtomic)
	skippedJobs := atomic.LoadUint64(skippedJobsAtomic)
	allJobs := addedJobs + skippedJobs
	duplicatePercentage := float64(0)
	if skippedJobs != 0 {
		duplicatePercentage = float64(skippedJobs) / float64(allJobs)
	}

	finishedJobsAtomic := &simulation.MetricsSystem.Metrics[metrics.JobsUnlockedCounter]
	finishedJobs := atomic.LoadUint64(finishedJobsAtomic)

	freeRespsAtomic := &simulation.MetricsSystem.Metrics[metrics.RespondersFreeCounter]
	busyRespsAtomic := &simulation.MetricsSystem.Metrics[metrics.RespondersBusyCounter]
	freeResps := atomic.LoadUint64(freeRespsAtomic)
	busyResps := atomic.LoadUint64(busyRespsAtomic)
	allResps := freeResps + busyResps
	loadPercentage := float64(0)
	if allResps != 0 {
		loadPercentage = float64(busyResps) / float64(allResps)
	}

	timeAverage := float64(0)
	if simulation.Pool.SpentTimeInPool != 0 {
		timeAverage = simulation.Pool.SpentTimeInPool / float64(simulation.Pool.PoppedAmount)
	}

	writer := tabwriter.NewWriter(os.Stdout, 48, 1, 1, ' ', 0)
	fmt.Fprintf(writer, "%s\n", "Общая статистика:")
	DrawValue(writer, "Количество обновлений", simulation.TickCounter)

	fmt.Fprintf(writer, "%s\n", "Тревоги:")
	DrawValue(writer, "Количество сохраннёных тревог", allAlerts)
	DrawValue(writer, "Количество перезаписанных тревог", rewrittenAlerts)
	DrawPercentage(writer, "Процент перезаписанных", rewritePercentage)

	fmt.Fprintf(writer, "%s\n", "Задачи:")
	DrawValue(writer, "Количество созданных задач", allJobs)
	DrawValue(writer, "Количество задач-дупликатов", skippedJobs)
	DrawPercentage(writer, "Процент дупликатов", duplicatePercentage)

	fmt.Fprintf(writer, "%s\n", "Обработка задач:")
	DrawValue(writer, "Количество завершенных задач", finishedJobs)
	DrawPercentage(writer, "Процент нагрузки", loadPercentage)
	DrawSeconds(writer, "Среднее время пребывания в системе", timeAverage)
	writer.Flush()

	fmt.Fprint(os.Stdout, "\n")

	ids := simulation.AgentsSystem.AgentsIds
	timesSpentInPool := make([]float64, len(ids))
	gotTimesSpentInPool := make([]bool, len(ids))
	timesSpentInPool, gotTimesSpentInPool = sparsemap.GetFromSparseMap(simulation.Pool.TimeLocked, timesSpentInPool, gotTimesSpentInPool, ids...)
	timesSpentHandling := make([]float64, len(ids))
	gotTimesSpentHandling := make([]bool, len(ids))
	timesSpentHandling, gotTimesSpentHandling = sparsemap.GetFromSparseMap(simulation.Pool.TimeUnlocked, timesSpentHandling, gotTimesSpentHandling, ids...)

	sources := tabwriter.NewWriter(os.Stdout, 16, 1, 1, ' ', 0)
	fmt.Fprintf(sources, "%s\n", "Статистика по источникам:")
	fmt.Fprintf(sources, "%s\t%s\t%s\t%s\t%s\n", "ID", "Создано", "Перезаписанно", "T БП", "T Обсл")
	for _, id := range ids {
		created := simulation.AgentsSystem.Created[id]
		rewritten := simulation.Buffer.Rewritten[id]
		timeSpentInPool := timesSpentInPool[id]
		timeSpentHandling := timesSpentHandling[id]

		fmt.Fprintf(sources, "%d\t%d\t%d\t%.2f\t%.2f\n",
			id,
			created,
			rewritten,
			timeSpentInPool,
			timeSpentHandling,
		)
	}
	sources.Flush()

	fmt.Fprint(os.Stdout, "\n")

	idsHandlers := simulation.RespondersSystem.Responders
	timesHandlersSpentHandling := make([]float64, len(idsHandlers))
	gotHandlersTimesSpentHandling := make([]bool, len(idsHandlers))
	timesHandlersSpentHandling, gotHandlersTimesSpentHandling = sparsemap.GetFromSparseMap(
		simulation.RespondersSystem.TimeUnlocked,
		timesHandlersSpentHandling,
		gotHandlersTimesSpentHandling,
		idsHandlers...,
	)

	handlers := tabwriter.NewWriter(os.Stdout, 16, 1, 1, ' ', 0)
	fmt.Fprintf(handlers, "%s\n", "Статистика по приборам:")
	fmt.Fprintf(handlers, "%s\t%s\t%s\n", "ID", "P Обсл", "T Обсл")
	for _, id := range idsHandlers {
		// all := simulation.RespondersSystem.All[id]
		handled := simulation.RespondersSystem.Handled[id]
		percentage := float64(0)
		if handled != 0 && finishedJobs != 0 {
			percentage = float64(handled) / float64(finishedJobs)
		}
		timeSpentHandling := timesHandlersSpentHandling[id]

		fmt.Fprintf(handlers, "%d\t%.2f\t%.2f\t\n",
			id,
			percentage,
			timeSpentHandling,
		)
	}
	handlers.Flush()
}

func DrawValue(writer *tabwriter.Writer, key string, value any) {
	fmt.Fprintf(writer, "%s:\t%v\n", key, value)
}

func DrawPercentage(writer *tabwriter.Writer, key string, value float64) {
	fmt.Fprintf(writer, "%s:\t%.2f\n", key, value)
}

func DrawSeconds(writer *tabwriter.Writer, key string, value float64) {
	fmt.Fprintf(writer, "%s:\t%.2f секунд\n", key, value)
}
