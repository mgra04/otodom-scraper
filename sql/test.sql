SELECT
    district,
    COUNT(*) AS apartment_count
FROM
    Apartmentsv2
GROUP BY
    district;