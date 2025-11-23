# Результаты нагрузочного тестирования

Тестирование проводилось утилитой k6 в Docker-окружении.

## Условия
*   **Инструмент:** k6
*   **Сценарий:** Создание команды, создание PR (транзакция), чтение списка ревью.
*   **Нагрузка:** 50 VUs (виртуальных пользователей), длительность 50s.

## Сводка результатов

| Метрика | Требование (SLI) | Результат теста |
| :--- | :--- | :--- |
| **Время ответа (p95)** | < 300 ms | 117.6 ms |
| **RPS** | 5 | ~67 |
| **Ошибки** | < 0.1% | 0.00% |

## Лог выполнения

```text
  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(95)<300' p(95)=117.6ms

    http_req_failed
    ✓ 'rate<0.01' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 3384    67.156295/s
    checks_succeeded...: 100.00% 3384 out of 3384
    checks_failed......: 0.00%   0 out of 3384

    ✓ team created
    ✓ pr created
    ✓ get reviews

    HTTP
    http_req_duration..............: avg=25.36ms min=1.16ms med=5.24ms max=350.44ms p(90)=75.15ms p(95)=117.6ms
      { expected_response:true }...: avg=25.36ms min=1.16ms med=5.24ms max=350.44ms p(90)=75.15ms p(95)=117.6ms
    http_req_failed................: 0.00%  0 out of 3384
    http_reqs......................: 3384   67.156295/s

    EXECUTION
    iteration_duration.............: avg=1.07s   min=1s     med=1.01s  max=1.7s     p(90)=1.23s   p(95)=1.31s  
    iterations.....................: 1128   22.385432/s
    vus............................: 3      min=1         max=49
    vus_max........................: 50     min=50        max=50

    NETWORK
    data_received..................: 57 MB  1.1 MB/s
    data_sent......................: 812 kB 16 kB/s
```