package task_manager

import (
	"cloudiac/common"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/runner"
	"context"
)

// 任务的流程步骤执行结束后的操作
// currStep 为 nil 表示未执行任何步骤
func (m *TaskManager) runTaskStepsDoneActions(ctx context.Context, task *models.Task, currStep *models.TaskStep, runTaskReq runner.RunTaskReq) error {
	logger := m.logger.WithField("func", "runTaskStepsDoneActions").WithField("taskId", task.Id)

	// 执行 callback 步骤
	func() {
		if currStep == nil {
			return
		}

		taskLastStep, err := services.GetTaskLastStep(m.db, task.Id)
		if err != nil {
			logger.Errorf("get task last step: %v", err)
			return
		}

		pipelineSteps := make([]models.PipelineStep, 0)
		if currStep.IsSuccess() && task.Flow.OnSuccess != nil {
			pipelineSteps = append(pipelineSteps, *task.Flow.OnSuccess)
		}

		if currStep.IsFail() && task.Flow.OnFail != nil {
			pipelineSteps = append(pipelineSteps, *task.Flow.OnFail)
		}

		newStepIndex := taskLastStep.Index + 1
		for _, pipelineStep := range pipelineSteps {
			step, err := services.CreateTaskCallbackStep(m.db, *task, pipelineStep, newStepIndex)
			if err != nil {
				logger.Errorf("create task callback step: %v", err)
				continue
			}
			newStepIndex += 1

			logger.Infof("run task callback step: %s", step.Id)
			if err := m.runTaskStep(ctx, runTaskReq, task, step); err != nil {
				logger.Infof("run task callback step: %v", err)
			}
		}
	}()

	if task.IsEffectTask() && currStep != nil && !currStep.IsRejected() {
		// 执行信息采集步骤
		logger.Infof("run task collect step")
		if err := m.runTaskStep(ctx, runTaskReq, task, &models.TaskStep{
			PipelineStep: models.PipelineStep{
				Type: models.TaskStepCollect,
			},
			OrgId:     task.OrgId,
			ProjectId: task.ProjectId,
			EnvId:     task.EnvId,
			TaskId:    task.Id,
			Index:     common.CollectTaskStepIndex,
			Status:    models.TaskStepPending,
		}); err != nil {
			logger.Errorf("run task collect step: %v", err)
		}
	}

	return nil
}
