import requests
import time
from prometheus_client.parser import text_string_to_metric_families


def parse_selector(selector):
    # If '{' is not present in the selector, no labels are being used
    if '{' not in selector:
        return selector, {}

    # Extracting metric name and label selectors from the PromQL selector
    metric_name, label_selectors_str = selector.split('{', 1)
    label_selectors_str = label_selectors_str.rstrip('}')

    # If label_selectors_str is empty, no labels are being used
    if not label_selectors_str:
        return metric_name, {}

    label_selectors = dict(s.strip().split('=')
                           for s in label_selectors_str.split(','))

    # Stripping quotes from label values
    label_selectors = {k: v.strip('\"') for k, v in label_selectors.items()}

    return metric_name, label_selectors


class PrometheusLibrary:
    def __init__(self, url):
        self.url = url


class PrometheusLibrary:
    def __init__(self, url):
        self.url = url

    def poll_and_parse(self):
        data = None
        for _ in range(10*4):
            try:
                response = requests.get(self.url)
                response.raise_for_status()  # Raise an exception if the GET request was unsuccessful
                data = response.text
                break
            except requests.exceptions.RequestException:
                time.sleep(0.25)
        if not data:
            raise RuntimeError(
                'Unable to retrieve data from the Prometheus exporter after 10 attempts')

        metrics = {}
        for family in text_string_to_metric_families(data):
            for sample in family.samples:
                metric_name = sample[0]
                metric_labels = sample[1]
                metric_value = sample[2]
                if metric_name not in metrics:
                    metrics[metric_name] = []
                metrics[metric_name].append({
                    "labels": metric_labels,
                    "value": metric_value
                })
        self.metrics = metrics
        return metrics

    def get_metric_by_selector(self, selector):
        metric_name, label_selectors = parse_selector(selector)
        if self.metrics is None:
            self.poll_and_parse()
        metrics = self.metrics

        matching_metrics = []
        if metric_name in metrics:
            for metric in metrics[metric_name]:
                labels = metric["labels"]
                if all(k in labels and str(labels[k]) == v for k, v in label_selectors.items()):
                    matching_metrics.append(metric)

        return matching_metrics

    def expect_metric_by_selector(self, selector, expected_value):
        metrics = self.get_metric_by_selector(selector)
        if not metrics:
            raise ValueError(f"No metrics matched the selector: {selector}")

        first_metric_value = metrics[0]['value']
        if first_metric_value != float(expected_value):
            raise ValueError(
                f"Expected value {expected_value}, but got {first_metric_value}")

        return True  # return True if the value matches the expected value
