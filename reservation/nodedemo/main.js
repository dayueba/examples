const { Client } = require('pg')

async function main() {
    const client = new Client({
      user: 'postgres',
      password: 'password',
      database: 'rsvp'
    })
    await client.connect()

    const res = await client.query(`select isempty((select during from reservation where teacher_id = 1108 ) * tsrange('["2010-01-01 14:20", "2010-01-01 14:25")'));`)
    console.log(res.rows)



    await client.end()
}

main()
