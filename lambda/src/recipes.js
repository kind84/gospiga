import * as redisearch from 'redredisearch'
import redis from 'redis'

async function recipes({args, dql}) {
    let uids = []
    let query = ""
    let vars = undefined
    if (args.query) {
        // ask redisearch
        const client = redis.createClient()
        redisearch.setClient(client)

        redisearch.createSearch('recipes', {}, function(err, search) {
            search
                .query(args.query)
                .end(function(err, ids) {
                    if (err) throw err
                    console.log(ids)
                    uids = ids
                })
        })
    }
    if (ids) {
        query = `query recipes($uid: [uid]) {
            recipes(func: type(Recipe)) @filter(uid) {
            }
        }`
        vars = {"$uids": ids}
    } else {
        query = `query recipes {
            recipes(func: type(Recipe)) @filter(eq(Author.name, $name)) {
            }
        }`
    }
    const results = await dql.query(query, vars)
    return results.data.recipes
}

self.addGraphQLResolvers({
    "Query.recipes": recipes,
})

