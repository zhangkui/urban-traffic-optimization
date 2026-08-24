import json, subprocess
from pathlib import Path
root=Path(__file__).resolve().parents[1]
fail=[]; rows=[]
def git(*args):
    return subprocess.run(['git',*args],cwd=root,text=True,capture_output=True).returncode == 0
for n in range(1,26):
    d=root/'cases'/f'BUG-{n:03d}'
    p=d/'metadata.json'
    packet=root/f'urban-traffic-optimization-bug-{n:03d}.md'
    if not p.exists(): fail.append(f'{n:03d}: metadata missing'); continue
    j=json.loads(p.read_text(encoding='utf-8-sig'))
    for key in ('bug_id','task_type','bug_type','green_branch','red_branch','g1_sha','red_sha','verify_cmds','gold_root_cause','success_criteria'):
        if not j.get(key): fail.append(f'{n:03d}: field {key} missing')
    if not packet.exists(): fail.append(f'{n:03d}: packet missing')
    if not git('show-ref','--verify',f'refs/heads/{j.get("green_branch")}'): fail.append(f'{n:03d}: green branch missing')
    if not git('show-ref','--verify',f'refs/heads/{j.get("red_branch")}'): fail.append(f'{n:03d}: red branch missing')
    rows.append(j)
counts={k:sum(1 for j in rows if j.get('task_type')==k) for k in ('bugfix','diagnosis')}
types={k:sum(1 for j in rows if j.get('bug_type')==k) for k in ('concurrency','nil','slice','error','context')}
if counts != {'bugfix':15,'diagnosis':10}: fail.append(f'task distribution {counts}')
if types != {k:5 for k in types}: fail.append(f'bug distribution {types}')
for name in ('README.md','openapi.yaml','docker-compose.yml','bug_summary.txt','bug_summary.xlsx','fix_check_report.md','selfcheck_create_report.md'):
    if not (root/name).exists(): fail.append(f'project artifact missing: {name}')
print('rows=',len(rows),'task_counts=',counts,'type_counts=',types)
print('FAIL=',len(fail))
for item in fail: print(' -',item)
(root/'local_stage8_checklist.md').write_text('# Local Stage8 Checklist\n\n- Records: %d\n- Task distribution: %s\n- Bug distribution: %s\n- Result: %s\n' % (len(rows),counts,types,'PASS' if not fail else 'FAIL') + ''.join('\n- '+x for x in fail),encoding='utf-8')
raise SystemExit(1 if fail else 0)
