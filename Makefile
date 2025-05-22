run_plan_debug:
	clear && go install . && TF_LOG=DEBUG terraform plan

run_plan_info:
	clear && go install . && TF_LOG=INFO terraform plan

run_apply_debug:
	clear && go install . && TF_LOG=DEBUG terraform apply

run_apply_info:
	clear && go install . && TF_LOG=INFO terraform apply