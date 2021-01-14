import * as redisearch from 'redredisearch'
import redis from 'redis'

async function recipes({args, dql}) {
    let uids = []
    let first = args.first ? `, ${args.first}` : ""
    let after = args.after ? `, ${args.after}` : ""

    if (args.query) {
        // ask redisearch
        const client = redis.createClient({ host: "redis" })
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

    let filterIDs = uids ? ", @filter(uid($uids))" : ""
    let query = `
        query Recipes($uids: string, $ingredients: string, $tags: string){
            recipes(
                func: type("Recipe")${first}${after}${filterIDs}
            ) {
                uid
                xid
                title
                subtitle
                mainImage {
                    uid
                    url
                }
                likes
                difficulty
                cost
                prepTime
                cookTime
                servings
                extraNotes
                description
                ingredients @filter(anyofterms(name, $ingredients)) {
                    uid
                    name
                    quantity
                    unitOfMeasure
                    food {
                        uid
                        term
                        stem
                    }
                }
                steps {
                    uid
                    heading
                    body
                    image {
                        uid
                        url
                    }
                }
                tags @filter(anyofterms(tagName, $tags)){
                    uid
                    tagName
                }
                conclusion
                slug
                createdAt
                modifiedAt
            }
        }
    `

    const vars = {"$uids": uids, "$ingredients": args.ingredients, "$tags": args.tags}

    const results = await dql.query(query, vars)

    return results.data.recipes
}

self.addGraphQLResolvers({
    "Query.recipes": recipes,
})

