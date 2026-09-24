#!/usr/bin/env python3
"""Exercise all services through Kong; assert real persisted checkout outcomes."""
import argparse,json,time,urllib.request,urllib.error,secrets,pathlib
p=argparse.ArgumentParser();p.add_argument('--rounds',type=int,default=3);args=p.parse_args()
BASE='http://127.0.0.1:8000/api/v1';results=[]
def call(method,path,data=None,expected=200):
 req=urllib.request.Request(BASE+path,data=json.dumps(data).encode() if data is not None else None,headers={'Content-Type':'application/json'},method=method)
 start=time.monotonic()
 try:r=urllib.request.urlopen(req,timeout=15)
 except urllib.error.HTTPError as e:r=e
 body=json.load(r);elapsed=round((time.monotonic()-start)*1000)
 assert r.status==expected,(path,r.status,body)
 trace=r.headers.get('X-Trace-ID');print(f'{method:5} {path:28} {r.status} {elapsed:4}ms trace={trace}')
 return body,trace,elapsed
for n in range(args.rounds):
 phone='08'+str(int(time.time()*1000))[-10:]+str(n%10)
 call('POST','/auth/register',{'name':'Demo Customer','phone_number':phone,'password':'demo-password','role':'customer'},201)
 login,_,_=call('POST','/auth/login',{'phone_number':phone,'password':'demo-password'})
 uid=login['user_id'];call('GET',f'/users/{uid}');call('GET',f'/wallets/{uid}');call('POST','/wallets/topup',{'user_id':uid,'amount':100000})
 payload={'user_id':uid,'origin_lat':-6.2,'origin_lng':106.8,'destination_lat':-6.3,'destination_lng':106.9,'fare':25000}
 for scenario,extra,expected in [('success',{},201),('slow',{'payment_delay_ms':1200},201),('insufficient-funds',{'fare':999999},500)]:
  body,trace,elapsed=call('POST','/orders',dict(payload,**extra),expected)
  if expected==201:
   order,_,_=call('GET',f'/orders/{body["id"]}');assert order['status']=='paid';assert order['user_name']=='Demo Customer'
  # Network scheduling and JSON decoding add jitter around the configured
  # 1200 ms server delay; assert the request was observably slow instead of
  # requiring an exact wall-clock boundary.
  if scenario=='slow':assert elapsed>=1000
  results.append({'scenario':scenario,'trace_id':trace,'duration_ms':elapsed,'response':body})
 wallet,_,_=call('GET',f'/wallets/{uid}');assert wallet['balance']==50000,wallet
 call('POST','/wallets/pay',{'user_id':uid,'amount':-1,'description':'invalid'},400)
 call('POST','/orders',dict(payload,user_id=999999999),500)
 print(f'Round {n+1}: persisted paid orders and wallet balance verified.');time.sleep(1)
pathlib.Path('artifacts').mkdir(exist_ok=True)
pathlib.Path('artifacts/simulation.json').write_text(json.dumps(results,indent=2))
print('PASS: all scenarios; trace IDs saved to artifacts/simulation.json')
