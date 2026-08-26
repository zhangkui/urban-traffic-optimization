INSERT INTO roads(id,name,code,level,speed_limit,lanes,length_meters,status) VALUES ('road-demo-1','人民路','RM001','arterial',50,4,3200,'active') ON CONFLICT DO NOTHING;
INSERT INTO intersections(id,name,code,latitude,longitude,level,status) VALUES ('intersection-demo-1','人民路-解放路口','INT001',31.2304,121.4737,'primary','online') ON CONFLICT DO NOTHING;
