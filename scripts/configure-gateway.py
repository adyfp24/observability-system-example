#!/usr/bin/env python3
"""Idempotently configure Kong's local Postgres-backed gateway."""
import json, urllib.request, urllib.error, time
BASE='http://127.0.0.1:8001'
def call(method,path,data=None):
 req=urllib.request.Request(BASE+path,data=json.dumps(data).encode() if data is not None else None,headers={'Content-Type':'application/json'},method=method)
 with urllib.request.urlopen(req,timeout=10) as r:return json.load(r)
for attempt in range(60):
 try:call('GET','/status');break
 except (OSError,urllib.error.URLError):time.sleep(2)
else:raise SystemExit('Kong did not become ready')
for name,paths in [('auth-user-service',['/api/v1/auth','/api/v1/users']),('order-service',['/api/v1/orders']),('payment-service',['/api/v1/wallets'])]:
 call('PUT','/services/'+name,{'name':name,'url':f'http://{name}:8080'})
 call('PUT','/routes/'+name,{'name':name,'service':{'name':name},'paths':paths,'strip_path':False})
existing=call('GET','/plugins')['data']
for name,config in [('cors',{'origins':['*'],'methods':['GET','POST','PUT','PATCH','DELETE','OPTIONS'],'headers':['*'],'credentials':True}),('opentelemetry',{'traces_endpoint':'http://otel-collector:4318/v1/traces','resource_attributes':{'service.name':'api-gateway'},'propagation':{'default_format':'w3c'}}),('prometheus',{'status_code_metrics':True,'latency_metrics':True,'bandwidth_metrics':True})]:
 plugin=next((p for p in existing if p['name']==name and not p.get('service') and not p.get('route')),None)
 call('PATCH' if plugin else 'POST','/plugins/'+plugin['id'] if plugin else '/plugins',{'name':name,'config':config})
print('Gateway routes, tracing, and metrics configured.')
