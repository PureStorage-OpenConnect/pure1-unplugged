#!/usr/bin/env sh
# Adds the necessary Kibana files into ElasticSearch to set up visualizations/dashboards

#
# To run an adhoc-test against a minikube Pure1 Unplugged instance (or to restore all of the kibana settings)
# 1. First enable port-forwarding to the elasticsearch cluster:
#        kubectl port-forward svc/pure1-unplugged-elasticsearch-client 9200:9200
#        kubectl port-forward svc/pure1-unplugged-kibana 5601:443
#    This will allow you to access elasticsearch on 127.0.0.1:9200
# 2. Next, go ahead to delete all of the dashboards and visualizations on Kibana. 
# 3. Finally inside this folder, run
#        PURE1_ES=127.0.0.1:9200 PURE1_KIBANA=127.0.0.1:5601 KIBANA_FILES=. ./kibana_setup.sh
#
set -ex

if [ -z "$PURE1_ES" ]; then
    PURE1_ES=pure1-unplugged-elasticsearch-client:9200
fi

if [ -z "$KIBANA_FILES" ]; then
    KIBANA_FILES=/kibana-files
fi

if [ -z "$PURE1_KIBANA" ]; then
    PURE1_KIBANA=pure1-unplugged-kibana:443
fi

# Wait until the first command succeeds
SETUP_SUCCESSFUL=false

# We wait for .kibana index to show up as a signal that Kibana has finished its bootstrapping
# Once it is ready we will proceed to set up all the configs, visualizations and dashboards.
while [ "$SETUP_SUCCESSFUL" != "true" ]
do
    STATUS_CODE=$(curl --max-time 30 -XGET -w "%{http_code}" $PURE1_ES/.kibana/_search?q=type:config --silent --output /dev/null)
    if [ "$STATUS_CODE" = 200 ];
    then
        SETUP_SUCCESSFUL=true
    else
        sleep 5
    fi
done

KIBANA_INDEX_PATTERNS="$KIBANA_FILES/index-patterns"
KIBANA_VISUALIZATIONS="$KIBANA_FILES/visualizations"
KIBANA_DASHBOARDS="$KIBANA_FILES/dashboards"

# Create an empty index in elasticsearch so that Kibana can create the index-pattern without reporting errors
curl -XPUT -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/pure-arrays-empty --data "@$KIBANA_FILES/pure_arrays_empty_index.json"

# Index patterns
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/index-pattern:pure_arrays_all --data "@$KIBANA_INDEX_PATTERNS/arrays.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/index-pattern:pure_arrays_metrics_all --data "@$KIBANA_INDEX_PATTERNS/arrays_metrics.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/index-pattern:pure_volumes_metrics_all --data "@$KIBANA_INDEX_PATTERNS/volumes_metrics.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/index-pattern:log_error_all --data "@$KIBANA_INDEX_PATTERNS/log_error.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/index-pattern:log_timer_all --data "@$KIBANA_INDEX_PATTERNS/log_timer.json"

# Main dashboard visualizations
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_main_control_viz --data "@$KIBANA_VISUALIZATIONS/array_main_control.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_main_object_summary_viz --data "@$KIBANA_VISUALIZATIONS/array_main_object_summary.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_main_capacity_summary_viz --data "@$KIBANA_VISUALIZATIONS/array_main_capacity_summary.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_main_performance_summary_viz --data "@$KIBANA_VISUALIZATIONS/array_main_performance_summary.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_read_latency_heatmap_viz --data "@$KIBANA_VISUALIZATIONS/array_read_latency_heatmap.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_write_latency_heatmap_viz --data "@$KIBANA_VISUALIZATIONS/array_write_latency_heatmap.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_percent_full_gauge_viz --data "@$KIBANA_VISUALIZATIONS/array_percent_full_gauge.json"

# Array performance visualizations
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_performance_control_viz --data "@$KIBANA_VISUALIZATIONS/array_performance_control.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_performance_summary_viz --data "@$KIBANA_VISUALIZATIONS/array_performance_summary.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_read_bandwidth_viz --data "@$KIBANA_VISUALIZATIONS/array_read_bandwidth.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_read_iops_viz --data "@$KIBANA_VISUALIZATIONS/array_read_iops.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_read_latency_viz --data "@$KIBANA_VISUALIZATIONS/array_read_latency.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_write_bandwidth_viz --data "@$KIBANA_VISUALIZATIONS/array_write_bandwidth.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_write_iops_viz --data "@$KIBANA_VISUALIZATIONS/array_write_iops.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_write_latency_viz --data "@$KIBANA_VISUALIZATIONS/array_write_latency.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_other_iops_viz --data "@$KIBANA_VISUALIZATIONS/array_other_iops.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_other_latency_viz --data "@$KIBANA_VISUALIZATIONS/array_other_latency.json"

# Volume performance visualizations
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_performance_control_viz --data "@$KIBANA_VISUALIZATIONS/volume_performance_control.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_performance_summary_viz --data "@$KIBANA_VISUALIZATIONS/volume_performance_summary.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_read_latency_viz --data "@$KIBANA_VISUALIZATIONS/volume_read_latency.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_write_latency_viz --data "@$KIBANA_VISUALIZATIONS/volume_write_latency.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_read_iops_viz --data "@$KIBANA_VISUALIZATIONS/volume_read_iops.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_write_iops_viz --data "@$KIBANA_VISUALIZATIONS/volume_write_iops.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_read_bandwidth_viz --data "@$KIBANA_VISUALIZATIONS/volume_read_bandwidth.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_write_bandwidth_viz --data "@$KIBANA_VISUALIZATIONS/volume_write_bandwidth.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_other_latency_viz --data "@$KIBANA_VISUALIZATIONS/volume_other_latency.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:volume_other_iops_viz --data "@$KIBANA_VISUALIZATIONS/volume_other_iops.json"

# File system performance visualizations (unique to file systems)
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:filesystem_performance_control_viz --data "@$KIBANA_VISUALIZATIONS/filesystem_performance_control.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:filesystem_performance_summary_viz --data "@$KIBANA_VISUALIZATIONS/filesystem_performance_summary.json"

# Array capacity visualizations
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_capacity_control_viz --data "@$KIBANA_VISUALIZATIONS/array_capacity_control.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_capacity_summary_viz --data "@$KIBANA_VISUALIZATIONS/array_capacity_summary.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_percent_full_viz --data "@$KIBANA_VISUALIZATIONS/array_percent_full.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_data_reduction_space_viz --data "@$KIBANA_VISUALIZATIONS/array_data_reduction.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_used_space_viz --data "@$KIBANA_VISUALIZATIONS/array_used_space.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_total_space_viz --data "@$KIBANA_VISUALIZATIONS/array_total_space.json"

# Log visualizations
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/search:log_errors_search --data "@$KIBANA_VISUALIZATIONS/log_errors.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/search:log_timers_search --data "@$KIBANA_VISUALIZATIONS/log_timers.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:log_error_levels_viz --data "@$KIBANA_VISUALIZATIONS/log_error_levels.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:log_error_sources_viz --data "@$KIBANA_VISUALIZATIONS/log_error_sources.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:log_timer_processes_viz --data "@$KIBANA_VISUALIZATIONS/log_timer_processes.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:log_timer_runtimes_viz --data "@$KIBANA_VISUALIZATIONS/log_timer_runtimes.json"

# Extra visualizations (not shown by default)
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/visualization:array_daily_used_space_viz --data "@$KIBANA_VISUALIZATIONS/array_daily_used_space.json"

# Predefined dashboards using the above visualizations
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/dashboard:main_dash --data "@$KIBANA_DASHBOARDS/main.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/dashboard:array_performance_dash --data "@$KIBANA_DASHBOARDS/array_performance.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/dashboard:volume_performance_dash --data "@$KIBANA_DASHBOARDS/volume_performance.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/dashboard:filesystem_performance_dash --data "@$KIBANA_DASHBOARDS/filesystem_performance.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/dashboard:array_capacity_dash --data "@$KIBANA_DASHBOARDS/array_capacity.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/dashboard:log_dash --data "@$KIBANA_DASHBOARDS/log.json"
curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/dashboard:array_daily_used_dash --data "@$KIBANA_DASHBOARDS/array_daily_used.json"

# Update kibana's config
# The config is only initiated when someone browse to the Kibana portal for the first time...
CONFIG_UPDATED=false
while [ "$CONFIG_UPDATED" != "true" ]
do
    # NOTE: Navigate to kibana home page once to trigger kibana to create the config object
    RESPONSE=$(curl -k -XGET $PURE1_KIBANA/app/kibana --silent --output /dev/null)

    # Now we try to update the config with custom config options (e.g. number formattings)
    RESPONSE=$(curl --max-time 10 -XGET $PURE1_ES/.kibana/_search?q=type:config --silent)
    CONFIG_ID=$(echo $RESPONSE | sed -n 's/.*"_id":"\([^"]*\)".*/\1/p')
    if [ -z "$CONFIG_ID" ];
     then
        sleep 5
     else
        curl -XPOST -H "Content-Type: application/json" -w "%{http_code}" $PURE1_ES/.kibana/doc/$CONFIG_ID/_update --data "@$KIBANA_FILES/kibana_config.json"
        CONFIG_UPDATED=true
     fi
done
