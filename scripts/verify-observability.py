#!/usr/bin/env python3
"""Assert the stored checkout trace, matching Loki logs, and Prometheus series."""
import json,time,urllib.request,urllib.parse,pathlib,base64

def get(url):
 with urllib.request.urlopen(url,timeout=10) as r:return json.load(r)

def verify():
 result=json.loads(pathlib.Path('artifacts/simulation.json').read_text())
 trace_id=next(r['trace_id'] for r in result if r['scenario']=='success')
 trace=get('http://127.0.0.1:3200/api/traces/'+trace_id)
 batches=trace.get('batches',trace.get('resourceSpans',[]));services=set();spans=[]
 for batch in batches:
  for a in batch['resource'].get('attributes',[]):
   if a['key']=='service.name':services.add(a['value']['stringValue'])
  for scope in batch.get('scopeSpans',batch.get('instrumentationLibrarySpans',[])):spans.extend(scope['spans'])
 required={'api-gateway','order-service','auth-user-service','payment-service'}
 assert required<=services,('trace services',services)
 assert any(s['name']=='postgresql' for s in spans),'database spans missing'
 span_ids={s['spanId'] for s in spans}
 for s in spans:
  parent=s.get('parentSpanId')
  if parent and parent not in ('0000000000000000','AAAAAAAAAAA='):assert parent in span_ids,('disconnected span',s['name'])
 query='{service_name=~".+"} |= "'+trace_id+'"'
 logs=get('http://127.0.0.1:3100/loki/api/v1/query_range?'+urllib.parse.urlencode({'query':query,'since':'1h','limit':1000}))
 log_services={x['stream']['service_name'] for x in logs['data']['result']}
 assert {'order-service','auth-user-service','payment-service'}<=log_services,('logs missing',log_services)
 metrics=get('http://127.0.0.1:9090/api/v1/query?'+urllib.parse.urlencode({'query':'demo_http_requests_total'}))
 metric_services={x['metric']['service'] for x in metrics['data']['result']}
 assert {'order-service','auth-user-service','payment-service'}<=metric_services,('metrics missing',metric_services)
 targets=get('http://127.0.0.1:9090/api/v1/targets')['data']['activeTargets'];assert len(targets)==4 and all(t['health']=='up' for t in targets),targets
 req=urllib.request.Request('http://127.0.0.1:3000/api/dashboards/uid/order-observability',headers={'Authorization':'Basic '+base64.b64encode(b'admin:demo-observability').decode()})
 with urllib.request.urlopen(req,timeout=10) as r:assert json.load(r)['dashboard']['uid']=='order-observability'
 report={'trace_id':trace_id,'trace_services':sorted(services),'span_count':len(spans),'log_services':sorted(log_services),'metric_services':sorted(metric_services),'scrape_targets':'4/4 up','dashboard':'provisioned'}
 pathlib.Path('artifacts/verification.json').write_text(json.dumps(report,indent=2));print(json.dumps(report,indent=2))
for attempt in range(12):
 try:verify();break
 except Exception as e:
  if attempt==29:raise
  print('Waiting for telemetry:',str(e)[:250]);time.sleep(3)
else:raise SystemExit(1)
print('PASS: connected gateway/service/database trace, correlated logs, metrics, and dashboard.')
